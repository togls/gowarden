package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/togls/gowarden/auth"
	"github.com/togls/gowarden/config"
	"github.com/togls/gowarden/handler/response"
	"github.com/togls/gowarden/model"
	"github.com/togls/gowarden/pkg/crypto"
	"github.com/togls/gowarden/store"
)

type OrganizationHandler struct {
	users   store.User
	ciphers store.Cipher
	orgs    store.Organization
	cs      store.Collection
	ops     store.OrgPolicy
	uos     store.UserOrganization
	ucs     store.UserCollection
	is      store.Invitation

	auth *auth.Core
	cfgs *config.Core
}

func NewOrganizationHandler(
	users store.User,
	ciphers store.Cipher,
	orgs store.Organization,
	cs store.Collection,
	ops store.OrgPolicy,
	uos store.UserOrganization,
	ucs store.UserCollection,
	is store.Invitation,
	auth *auth.Core,
	cfgs *config.Core,
) *OrganizationHandler {
	return &OrganizationHandler{
		users:   users,
		ciphers: ciphers,
		orgs:    orgs,
		cs:      cs,
		ops:     ops,
		uos:     uos,
		ucs:     ucs,
		is:      is,
		auth:    auth,
		cfgs:    cfgs,
	}
}

func (oh *OrganizationHandler) Routes(e *echo.Echo) {
	e.GET("/api/ciphers/organization-details", oh.GetOrgDetails, oh.auth.RequireAuth)

	{
		cl := e.Group("/api/collections")
		cl.GET("", oh.GetUserCollections, oh.auth.RequireAuth)
	}

	{ // base auth
		org := e.Group("/api/organizations", oh.auth.RequireAuth)
		org.POST("", oh.CreateOrganization)
		org.POST("/:uuid/leave", oh.LeaveOrganization)
	}

	{ // owner
		org := e.Group("/api/organizations", oh.auth.RequireOwnerAuth)
		org.GET("/:uuid", oh.GetOrganization)
		org.DELETE(":uuid", oh.DeleteOrganization)
		org.POST("/:uuid/delete", oh.PostDeleteOrganization)
		org.PUT("/:uuid", oh.PutOrganization)
		org.POST("/:uuid", oh.PostOrganization)
	}

	{ // ManagerLoose
		org := e.Group("/api/organizations", oh.auth.RequireManagerLooseAuth)
		org.GET("/:uuid/collections", oh.GetOrgCollections)
		org.POST("/:uuid/collections", oh.PostOrganizationCollections)
		org.GET("/:ouuid/users", oh.GetOrgUsers)
	}

	{ // Manager
		org := e.Group("/api/organizations", oh.auth.RequireManagerAuth)
		org.GET("/:ouuid/collections/:cuuid/details", oh.GetOrgCollectionDetail)
		org.GET("/:ouuid/collections/:cuuid/users", oh.GetCollectionUsers)
		org.PUT("/:ouuid/collections/:cuuid/users", oh.PutCollectionUsers)
		org.POST("/:ouuid/collections/:cuuid", oh.PostOrganizationCollectionUpdate)
		org.PUT("/:ouuid/collections/:cuuid", oh.PutOrganizationCollectionUpdate)
		org.DELETE("/:ouuid/collections/:cuuid", oh.DeleteOrganizationCollection)
		org.POST("/:ouuid/collections/:cuuid/delete", oh.PostOrganizationCollectionDelete)
	}

	{ // require admin
		org := e.Group("/api/organizations", oh.auth.RequireAdminAuth)
		org.DELETE("/:ouuid/collections/:cuuid/users/:uouuid", oh.DeleteOrganizationCollectionUser)
		org.POST("/:ouuid/collections/:cuuid/delete-users/:uouuid", oh.PostOrganizationCollectionDeleteUser)
		org.POST("/:ouuid/users/invite", oh.SendInvite)
		org.GET("/:ouuid/users/:uouuid", oh.GetUser)
		org.POST("/:ouuid/users/:uouuid", oh.EditUser)
		org.PUT("/:ouuid/users/:uouuid", oh.PutUser)
		org.DELETE("/:ouuid/users/:uouuid", oh.DeleteUser)
		org.DELETE("/:ouuid/users", oh.BulkDeleteUser)
		org.POST("/:ouuid/users/:uouuid/delete", oh.PostDeleteUser)
	}

	// TODO:
	// reinvite_user
	// bulk_reinvite_user
	// confirm_invite
	// bulk_confirm_invite
	// accept_invite
	// post_org_import,
	// list_policies,
	// list_policies_token,
	// get_policy,
	// put_policy,
	// get_organization_tax,
	// get_plans,
	// get_plans_tax_rates,
	// import,
	// post_org_keys,
	// bulk_public_keys,
}

func (oh *OrganizationHandler) GetOrganization(c echo.Context) error {
	oUuid := c.Param("uuid")

	organization, err := oh.orgs.FindByUuid(oUuid)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Can't find organization details").SetInternal(err)
	}

	return c.JSON(http.StatusOK, organization)
}

type OrgData struct {
	BillingEmail   string      `json:"BillingEmail"`
	CollectionName string      `json:"CollectionName"`
	Key            *string     `json:"Key"`
	Name           string      `json:"Name"`
	Keys           *OrgKeyData `json:"Keys"`
}

type OrgKeyData struct {
	EncryptedPrivateKey string `json:"EncryptedPrivateKey"`
	PublicKey           string `json:"PublicKey"`
}

func (oh *OrganizationHandler) CreateOrganization(c echo.Context) error {
	data := new(OrgData)

	if err := c.Bind(data); err != nil {
		return err
	}

	user := auth.GetUser(c)

	user, err := oh.users.FindByUuid(user.Uuid)
	if err != nil {
		return err
	}

	if !oh.cfgs.IsOrgCreationAllowed(user.Email) {
		return echo.NewHTTPError(http.StatusBadRequest, "User not allowed to create organizations")
	}

	ps, err := oh.ops.FindConfirmedByUser(user.Uuid)
	if err != nil {
		return err
	}

	applicable := func() bool {
		for _, p := range ps {
			if p.Enabled && p.Atype == model.OPTypeSingleOrg {
				oUuid := p.OrgUuid
				uo, err := oh.uos.FindByUserAndOrg(user.Uuid, oUuid)
				if err != nil {
					return false
				}

				if uo.Atype == model.UOTypeAdmin {
					return true
				}
			}
		}

		return false
	}

	if applicable() {
		return echo.NewHTTPError(http.StatusBadRequest, "You may not create an organization. You belong to an organization which has a policy that prohibits you from being a member of any other organization.")
	}

	oUuid, err := crypto.GenerateUuid()
	if err != nil {
		return err
	}

	org := &model.Organization{
		Uuid:         oUuid,
		Name:         data.Name,
		BillingEmail: data.BillingEmail,
	}

	if data.Keys != nil {
		org.PublicKey = &data.Keys.PublicKey
		org.PrivateKey = &data.Keys.EncryptedPrivateKey
	}

	if err := oh.orgs.Create(org); err != nil {
		return err
	}

	uoUuid, err := crypto.GenerateUuid()
	if err != nil {
		return err
	}

	uo := &model.UserOrganization{
		Uuid:      uoUuid,
		UserUuid:  user.Uuid,
		OrgUuid:   oUuid,
		AccessAll: true,
		Status:    model.UOStatusConfirmed,
		Atype:     model.UOTypeOwner,
		AKey:      data.Key,
	}
	if err := oh.uos.Create(uo); err != nil {
		return err
	}

	cUuid, err := crypto.GenerateUuid()
	if err != nil {
		return err
	}
	cl := &model.Collection{
		Uuid:    cUuid,
		Name:    data.CollectionName,
		OrgUuid: oUuid,
	}
	if err := oh.cs.Save(cl); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, org)
}

func (oh *OrganizationHandler) DeleteOrganization(c echo.Context) error {
	data := new(PasswordData)

	if err := c.Bind(data); err != nil {
		return err
	}

	oUuid := c.Param("uuid")

	user := auth.GetUser(c)

	user, err := oh.users.FindByUuid(user.Uuid)
	if err != nil {
		return err
	}

	ok := crypto.VerifyPassword(
		data.MasterPasswordHash,
		user.Salt,
		user.PasswordHash,
		user.PasswordIterations,
	)
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid password")
	}

	org, err := oh.orgs.FindByUuid(oUuid)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Organization not found").SetInternal(err)
	}

	if err := oh.orgs.Delete(org.Uuid); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (oh *OrganizationHandler) PostDeleteOrganization(c echo.Context) error {
	return oh.DeleteOrganization(c)
}

func (oh *OrganizationHandler) LeaveOrganization(c echo.Context) error {
	oUuid := c.Param("uuid")

	user := auth.GetUser(c)

	uo, err := oh.uos.FindByUserAndOrg(user.Uuid, oUuid)
	if err != nil {
		return err
	}

	// delete users_collections
	if !uo.AccessAll {
		cs, err := oh.cs.Find(&model.CollectionFilter{OrgUuid: &oUuid})
		if err != nil {
			return err
		}

		IDs := make([]string, len(cs))
		for _, c := range cs {
			IDs = append(IDs, c.Uuid)
		}

		if err := oh.cs.DeleteUser(IDs, user.Uuid); err != nil {
			return err
		}
	}

	now := time.Now()
	uu := &model.UpdateUser{
		Uuid: user.Uuid,

		UpdatedAt: &now,
	}
	if err := oh.users.Update(uu); err != nil {
		return err
	}

	if err := oh.uos.Delete(uo.Uuid); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (oh *OrganizationHandler) GetUserCollections(c echo.Context) error {
	user := auth.GetUser(c)

	collections, err := oh.cs.Find(&model.CollectionFilter{UserUuid: &user.Uuid})
	if err != nil {
		return err
	}

	resp := &struct {
		Data              []*model.Collection `json:"Data"`
		Object            string              `json:"Object"`
		ContinuationToken any
	}{
		Data:              collections,
		Object:            "list",
		ContinuationToken: nil,
	}

	return c.JSON(http.StatusOK, resp)
}

func (oh *OrganizationHandler) GetOrgCollections(c echo.Context) error {
	oUuid := c.Param("uuid")

	collections, err := oh.cs.Find(&model.CollectionFilter{OrgUuid: &oUuid})
	if err != nil {
		return err
	}

	resp := &struct {
		Data              []*model.Collection `json:"Data"`
		Object            string              `json:"Object"`
		ContinuationToken any
	}{
		Data:              collections,
		Object:            "list",
		ContinuationToken: nil,
	}

	return c.JSON(http.StatusOK, resp)
}

func (oh *OrganizationHandler) GetOrgCollectionDetail(c echo.Context) error {
	oUuid := c.Param("ouuid")
	cUuid := c.Param("cuuid")

	user := auth.GetUser(c)

	collection, err := oh.cs.FindByCollectionUser(cUuid, user.Uuid)
	if err != nil {
		return err
	}

	if collection.OrgUuid != oUuid {
		return echo.NewHTTPError(http.StatusBadRequest, "Collection is not owned by organization")
	}

	return c.JSON(http.StatusOK, collection)
}

func (oh *OrganizationHandler) GetCollectionUsers(c echo.Context) error {
	oUuid := c.Param("ouuid")
	cUuid := c.Param("cuuid")

	collection, err := oh.cs.FindByCollectionOrg(cUuid, oUuid)
	if err != nil {
		return err
	}

	ucs, err := oh.ucs.Find(&model.UCFilter{CollectionUuid: &collection.Uuid})
	if err != nil {
		return err
	}

	type Resp struct {
		ID            string `json:"Id"`
		ReadOnly      bool   `json:"ReadOnly"`
		HidePasswords bool   `json:"HidePasswords"`
	}

	resp := make([]*Resp, 0)
	for _, cu := range ucs {
		uo, err := oh.uos.FindByUserAndOrg(cu.UserUuid, oUuid)
		if err != nil {
			return err
		}

		resp = append(resp, &Resp{
			ID:            uo.Uuid,
			ReadOnly:      cu.ReadOnly,
			HidePasswords: cu.HidePasswords,
		})
	}

	return c.JSON(http.StatusOK, resp)
}

type CollectionData struct {
	ID            string `json:"Id"`
	ReadOnly      bool   `json:"ReadOnly"`
	HidePasswords bool   `json:"HidePasswords"`
}

func (oh *OrganizationHandler) PutCollectionUsers(c echo.Context) error {
	oUuid := c.Param("ouuid")
	cUuid := c.Param("cuuid")

	data := new([]*CollectionData)

	if err := c.Bind(data); err != nil {
		return err
	}

	collection, err := oh.cs.FindByCollectionOrg(cUuid, oUuid)
	if err != nil || collection == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Collection not found in Organization")
	}

	// delete all
	if err := oh.ucs.DeleteAllByCollection(cUuid); err != nil {
		return err
	}

	// add new
	for _, v := range *data {
		uo, err := oh.uos.FindByUuid(v.ID)
		if err != nil || uo == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "User is not part of organization")
		}

		if uo.AccessAll {
			continue
		}

		if err := oh.ucs.Save(cUuid, uo.UserUuid, v.ReadOnly, v.HidePasswords); err != nil {
			return err
		}
	}

	return c.NoContent(http.StatusOK)
}

type EditUserData struct {
	Type        int               `json:"Type,string"`
	Collections []*CollectionData `json:"Collections"`
	AccessAll   bool              `json:"AccessAll"`
}

func (oh *OrganizationHandler) EditUser(c echo.Context) error {
	oUuid := c.Param("ouuid")
	uoUuid := c.Param("uouuid")
	data := new(EditUserData)

	userOrg := auth.GetUserOrganization(c)

	uoType := userOrg.Atype

	if err := c.Bind(data); err != nil {
		return err
	}

	editUO, err := oh.uos.FindByUuid(uoUuid)
	if err != nil || editUO.OrgUuid != oUuid {
		return echo.NewHTTPError(http.StatusBadRequest, "The specified user isn't member of the organization")
	}

	newType := model.UOType(data.Type)
	if newType != model.UOTypeUser &&
		(editUO.Atype >= model.UOTypeAdmin || newType >= model.UOTypeAdmin) &&
		(uoType != model.UOTypeOwner) {
		return echo.NewHTTPError(http.StatusBadRequest, "Only Owners can grant and remove Admin or Owner privileges")
	}

	if editUO.Atype == model.UOTypeOwner && uoType != model.UOTypeOwner {
		return echo.NewHTTPError(http.StatusBadRequest, "Only Owners can edit Owner users")
	}

	if editUO.Atype == model.UOTypeOwner && newType != model.UOTypeOwner {
		owner := model.UOTypeOwner
		uos, err := oh.uos.Find(&model.UOFilter{OrgUuid: &oUuid, Atype: &owner})
		if err != nil {
			return err
		}

		if len(uos) <= 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "Can't delete the last owner")
		}
	}

	editUO.AccessAll = data.AccessAll
	editUO.Atype = newType

	if err := oh.ucs.DeleteAllByUserAndOrg(editUO.UserUuid, oUuid); err != nil {
		return err
	}

	if !data.AccessAll {
		for _, cl := range data.Collections {
			collection, err := oh.cs.FindByCollectionOrg(cl.ID, oUuid)
			if err != nil {
				return err
			}

			if err := oh.ucs.Save(collection.Uuid, editUO.UserUuid, cl.ReadOnly, cl.HidePasswords); err != nil {
				return err
			}
		}
	}

	if err := oh.uos.Save(editUO); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (oh *OrganizationHandler) PutOrganization(c echo.Context) error {
	return oh.PostOrganization(c)
}

type OrganizationUpdateData struct {
	BillingEmail string `json:"BillingEmail"`
	Name         string `json:"Name"`
}

func (oh *OrganizationHandler) PostOrganization(c echo.Context) error {
	data := new(OrganizationUpdateData)

	if err := c.Bind(data); err != nil {
		return err
	}

	oUuid := c.Param("uuid")

	org, err := oh.orgs.FindByUuid(oUuid)
	if err != nil {
		return err
	}

	org.Name = data.Name
	org.BillingEmail = data.BillingEmail

	if err := oh.orgs.Save(org); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, org)
}

type NewCollectionData struct {
	Name string `json:"Name"`
}

func (oh *OrganizationHandler) PostOrganizationCollections(c echo.Context) error {
	data := new(NewCollectionData)

	if err := c.Bind(data); err != nil {
		return err
	}

	oUuid := c.Param("uuid")

	user := auth.GetUser(c)

	org, err := oh.orgs.FindByUuid(oUuid)
	if err != nil {
		return err
	}

	uo, err := oh.uos.FindByUserAndOrg(user.Uuid, org.Uuid)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "User is not part of organization")
	}

	cUuid, err := crypto.GenerateUuid()
	if err != nil {
		return err
	}

	cl := &model.Collection{
		Uuid:    cUuid,
		OrgUuid: org.Uuid,
		Name:    data.Name,
	}

	if err := oh.cs.Save(cl); err != nil {
		return err
	}

	if !uo.AccessAll {
		oh.ucs.Save(cUuid, uo.UserUuid, false, false)
	}

	return c.JSON(http.StatusOK, cl)
}

func (oh *OrganizationHandler) DeleteOrganizationCollectionUser(c echo.Context) error {
	oUuid := c.Param("ouuid")
	cUuid := c.Param("cuuid")
	uoUuid := c.Param("uouuid")

	cl, err := oh.cs.FindByUuid(cUuid)
	if err != nil {
		return err
	}

	if cl.OrgUuid != oUuid {
		return echo.NewHTTPError(http.StatusBadRequest, "Collection and Organization id do not match")
	}

	uo, err := oh.uos.FindByUuid(uoUuid)
	if err != nil {
		return err
	}

	if uo.OrgUuid != oUuid {
		return echo.NewHTTPError(http.StatusBadRequest, "User not found in organization")
	}

	cu, err := oh.ucs.FindByCollectionUser(cUuid, uo.UserUuid)
	if err != nil {
		return err
	}

	if err := oh.ucs.DeleteByUserCollection(cu.CollectionUuid, cu.UserUuid); err != nil {
		return err
	}

	if err := oh.users.UpdateRevision(uo.UserUuid); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (oh *OrganizationHandler) PostOrganizationCollectionDeleteUser(c echo.Context) error {
	return oh.DeleteOrganizationCollectionUser(c)
}

func (oh *OrganizationHandler) PostOrganizationCollectionUpdate(c echo.Context) error {
	data := new(NewCollectionData)

	if err := c.Bind(data); err != nil {
		return err
	}

	oUuid := c.Param("ouuid")
	cUuid := c.Param("cuuid")

	org, err := oh.orgs.FindByUuid(oUuid)
	if err != nil {
		return err
	}

	cl, err := oh.cs.FindByUuid(cUuid)
	if err != nil {
		return err
	}

	if cl.OrgUuid != org.Uuid {
		return echo.NewHTTPError(http.StatusBadRequest, "Collection is not owned by organization")
	}

	cl.Name = data.Name

	if err := oh.cs.Save(cl); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, cl)
}

func (oh *OrganizationHandler) PutOrganizationCollectionUpdate(c echo.Context) error {
	return oh.PostOrganizationCollectionUpdate(c)
}

func (oh *OrganizationHandler) DeleteOrganizationCollection(c echo.Context) error {
	oUuid := c.Param("ouuid")
	cUuid := c.Param("cuuid")

	org, err := oh.orgs.FindByUuid(oUuid)
	if err != nil {
		return err
	}

	cl, err := oh.cs.FindByUuid(cUuid)
	if err != nil {
		return err
	}

	if cl.OrgUuid != org.Uuid {
		return echo.NewHTTPError(http.StatusBadRequest, "Collection is not owned by organization")
	}

	if err := oh.cs.Delete(cUuid); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (oh *OrganizationHandler) PostOrganizationCollectionDelete(c echo.Context) error {
	return oh.DeleteOrganizationCollection(c)
}

type OrgIdData struct {
	OrganizationId string `query:"organizationId" form:"organizationId"`
}

func (oh *OrganizationHandler) GetOrgDetails(c echo.Context) error {
	data := new(OrgIdData)

	if err := c.Bind(data); err != nil {
		return err
	}

	ciphers, err := oh.ciphers.FindByOrg(data.OrganizationId)
	if err != nil {
		return err
	}

	var cs []*response.Cipher
	for i := range ciphers {
		rc, err := response.NewCipher(ciphers[i])
		if err != nil {
			return err
		}

		cs = append(cs, rc)
	}

	resp := &RespListData{
		Data:              cs,
		Object:            "list",
		ContinuationToken: nil,
	}

	return c.JSON(http.StatusOK, resp)
}

type RespListData struct {
	Data              any    `json:"Data"`
	Object            string `json:"Object"`
	ContinuationToken any    `json:"ContinuationToken"`
}

func (oh *OrganizationHandler) GetOrgUsers(c echo.Context) error {
	oUuid := c.Param("ouuid")

	uos, err := oh.uos.Find(&model.UOFilter{OrgUuid: &oUuid})
	if err != nil {
		return err
	}

	resp := &RespListData{
		Data:              uos,
		Object:            "list",
		ContinuationToken: nil,
	}

	return c.JSON(http.StatusOK, resp)
}

type InviteData struct {
	Emails      []string         `json:"Emails"`
	Type        *model.UOType    `json:"Type"`
	Collections []CollectionData `json:"Collections"`
	AccessAll   *bool            `json:"AccessAll"`
}

func (id *InviteData) UnmarshalJSON(data []byte) error {
	type Alias InviteData
	alias := &struct {
		Type int `json:"Type"`
		*Alias
	}{
		Alias: (*Alias)(id),
	}

	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	t := model.UOType(alias.Type)

	id.Type = &t

	return nil
}

func (oh *OrganizationHandler) SendInvite(c echo.Context) error {
	data := new(InviteData)

	if err := c.Bind(data); err != nil {
		return err
	}

	oUuid := c.Param("ouuid")

	userOrg := auth.GetUserOrganization(c)
	userOrgType := userOrg.Atype

	newType := *data.Type

	if newType != model.UOTypeUser && userOrgType != model.UOTypeOwner {
		return echo.NewHTTPError(http.StatusForbidden, "Only Owners can invite Managers, Admins or Owners")
	}

	mailEnabled := false
	uoStatus := model.UOStatusAccepted
	if mailEnabled {
		uoStatus = model.UOStatusInvited
	}

	for _, email := range data.Emails {
		email := strings.ToLower(email)
		user, err := oh.users.FindByEmail(email)
		if err != nil && !errors.Is(err, model.ErrNotFound) {
			return err
		}

		if errors.Is(err, model.ErrNotFound) {
			if !oh.cfgs.IsInvitationsAllowed() {
				return echo.NewHTTPError(http.StatusForbidden, "Invitations are disabled")
			}

			if !oh.cfgs.IsEmailDomainAllowed(email) {
				return echo.NewHTTPError(http.StatusForbidden, "Email domain not eligible for invitations")
			}

			if !mailEnabled {
				if err := oh.is.Save(&model.Invitation{Email: email}); err != nil {
					return err
				}
			}

			user, err = oh.newUser(email)
			if err != nil {
				return err
			}
			if err := oh.users.Create(user); err != nil {
				return err
			}

			uoStatus = model.UOStatusInvited
		} else {
			uo, err := oh.uos.FindByUserAndOrg(user.Uuid, oUuid)
			if err != nil && !errors.Is(err, model.ErrNotFound) {
				return err
			}

			if uo != nil {
				return echo.NewHTTPError(http.StatusBadRequest, "User already in organization: "+email)
			}
		}

		id, err := crypto.GenerateUuid()
		if err != nil {
			return err
		}
		uo := &model.UserOrganization{
			Uuid:      id,
			UserUuid:  user.Uuid,
			OrgUuid:   oUuid,
			AccessAll: *data.AccessAll,
			Atype:     newType,
			Status:    uoStatus,
		}
		if err := oh.uos.Save(uo); err != nil {
			return err
		}

		if !*data.AccessAll {
			for _, c := range data.Collections {
				cl, err := oh.cs.FindByCollectionOrg(c.ID, oUuid)
				if err != nil {
					return err
				}

				if err := oh.ucs.Save(cl.Uuid, user.Uuid, c.ReadOnly, c.HidePasswords); err != nil {
					return err
				}
			}
		}

		// TODO: send email
	}

	return c.NoContent(http.StatusOK)
}

func (oh *OrganizationHandler) newUser(email string) (*model.User, error) {
	id, err := crypto.GenerateUuid()
	if err != nil {
		return nil, err
	}

	ss, err := crypto.GenerateUuid()
	if err != nil {
		return nil, err
	}

	salt, err := crypto.GenerateBytes(64)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	user := &model.User{
		Uuid:               id,
		Enabled:            true,
		CreatedAt:          now,
		UpdatedAt:          now,
		Email:              email,
		Name:               email,
		Salt:               salt,
		PasswordIterations: oh.cfgs.PasswordIterations,
		SecurityStamp:      ss,
		ClientKdfType:      model.ClientKdfTypeDefault,
		ClientKdfIter:      model.ClientKdfIterDefault,
	}

	return user, nil
}

func (oh *OrganizationHandler) GetUser(c echo.Context) error {
	oUuid := c.Param("ouuid")
	uoUuid := c.Param("uouuid")

	uo, err := oh.uos.FindByUuid(uoUuid)
	if err != nil {
		return err
	}

	if uo.OrgUuid != oUuid {
		return echo.NewHTTPError(http.StatusBadRequest, "The specified user isn't a member of the organization")
	}

	var cls []*model.UOCollection

	if !uo.AccessAll {
		cs, err := oh.ucs.Find(&model.UCFilter{UserUuid: &uo.UserUuid, OrgUuid: &oUuid})
		if err != nil {
			return err
		}

		for _, c := range cs {
			cls = append(cls, &model.UOCollection{
				ID:            c.CollectionUuid,
				ReadOnly:      c.ReadOnly,
				HidePasswords: c.HidePasswords,
			})
		}

	}

	d := &model.UODetail{
		ID:          uo.Uuid,
		UserID:      uo.UserUuid,
		Status:      int(uo.Status),
		Type:        int(uo.Atype),
		AccessAll:   uo.AccessAll,
		Collections: cls,
		Object:      "organizationUserDetails",
	}

	return c.JSON(http.StatusOK, d)
}

func (oh *OrganizationHandler) PutUser(c echo.Context) error {
	return oh.EditUser(c)
}

func (oh *OrganizationHandler) DeleteUser(c echo.Context) error {
	oUuid := c.Param("ouuid")
	uoUuid := c.Param("uouuid")

	userOrg := auth.GetUserOrganization(c)
	userOrgType := userOrg.Atype

	if err := oh.deleteUserOrg(uoUuid, oUuid, &userOrgType); err != nil {
		return err
	}

	return c.NoContent(http.StatusOK)
}

func (oh OrganizationHandler) deleteUserOrg(uoUuid, oUuid string, userOrgType *model.UOType) error {
	uo, err := oh.uos.FindByUuid(uoUuid)
	if err != nil {
		return err
	}

	if uo.OrgUuid != oUuid {
		return echo.NewHTTPError(http.StatusBadRequest, "User to delete isn't member of the organization")
	}

	if uo.Atype != model.UOTypeUser && *userOrgType != model.UOTypeOwner {
		return echo.NewHTTPError(http.StatusForbidden, "Only Owners can delete Admins or Owners")
	}

	if uo.Atype == model.UOTypeOwner {
		owner := model.UOTypeOwner
		uolist, err := oh.uos.Find(&model.UOFilter{OrgUuid: &oUuid, Atype: &owner})
		if err != nil {
			return err
		}

		if len(uolist) <= 1 {
			return echo.NewHTTPError(http.StatusBadRequest, "Can't delete the last owner")
		}
	}

	now := time.Now()
	uu := &model.UpdateUser{
		Uuid:      uo.UserUuid,
		UpdatedAt: &now,
	}
	if err := oh.users.Update(uu); err != nil {
		return err
	}

	if err := oh.ucs.DeleteAllByUserAndOrg(uo.UserUuid, oUuid); err != nil {
		return err
	}

	if err := oh.uos.Delete(uo.Uuid); err != nil {
		return err
	}

	return nil
}

type OrgIDsData struct {
	IDs []string `json:"Ids"`
}

func (oh *OrganizationHandler) BulkDeleteUser(c echo.Context) error {
	oUuid := c.Param("ouuid")

	userOrg := auth.GetUserOrganization(c)
	userOrgType := userOrg.Atype

	var data OrgIDsData
	if err := c.Bind(&data); err != nil {
		return err
	}

	type RespItem struct {
		Object string `json:"Object"`
		Id     string `json:"Id"`
		Error  string `json:"Error"`
	}
	var resp []RespItem
	for _, id := range data.IDs {
		msg := ""
		herr := echo.NewHTTPError(http.StatusBadRequest)
		err := oh.deleteUserOrg(id, oUuid, &userOrgType)
		if errors.As(err, &herr) {
			msg = fmt.Sprintf("%s", herr.Message)
		}

		resp = append(resp, RespItem{
			Object: "OrganizationBulkConfirmResponseModel",
			Id:     id,
			Error:  msg,
		})
	}

	return c.JSON(http.StatusOK, struct {
		Data              []RespItem `json:"Data"`
		Object            string     `json:"Object"`
		ContinuationToken any        `json:"ContinuationToken"`
	}{
		Data:              resp,
		Object:            "list",
		ContinuationToken: nil,
	})
}

func (oh *OrganizationHandler) PostDeleteUser(c echo.Context) error {
	return oh.DeleteUser(c)
}
