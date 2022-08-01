package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/togls/gowarden/model"
)

const (
	deviceKey           = "middleware-device"
	userKey             = "middleware-user"
	userOrganizationKey = "middleware-user-organization"
)

func GetUser(c echo.Context) *model.User {
	return c.Get(userKey).(*model.User)
}

func GetDevice(c echo.Context) *model.Device {
	return c.Get(deviceKey).(*model.Device)
}

func GetUserOrganization(c echo.Context) *model.UserOrganization {
	return c.Get(userOrganizationKey).(*model.UserOrganization)
}

func (core Core) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, device, err := core.baseAuth(c)
		if err != nil {
			return err
		}

		c.Set(userKey, user)
		c.Set(deviceKey, device)
		return next(c)
	}
}

func (core Core) RequireOrgAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, device, err := core.baseAuth(c)
		if err != nil {
			return err
		}

		uo, err := core.orgAuth(c, user)
		if err != nil {
			return err
		}

		c.Set(userKey, user)
		c.Set(deviceKey, device)
		c.Set(userOrganizationKey, uo)

		return next(c)
	}
}

func (core Core) RequireAdminAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, device, err := core.baseAuth(c)
		if err != nil {
			return err
		}

		uo, err := core.orgAuth(c, user)
		if err != nil {
			return err
		}

		if uo.Atype > model.UOTypeAdmin {
			return c.String(http.StatusUnauthorized, "The current user isn't admin of the organization")
		}

		c.Set(userKey, user)
		c.Set(deviceKey, device)
		c.Set(userOrganizationKey, uo)

		return next(c)
	}
}

// check org id and collection id
func (core Core) RequireManagerAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, device, err := core.baseAuth(c)
		if err != nil {
			return err
		}

		uo, err := core.orgAuth(c, user)
		if err != nil {
			return err
		}

		if uo.Atype > model.UOTypeAdmin && uo.Atype != model.UOTypeManager {
			return c.String(http.StatusUnauthorized,
				"You need to be a Manager, Admin or Owner to call this endpoint")
		}

		cUuid := getCollectionUuid(c)
		if cUuid == "" {
			return c.String(http.StatusUnauthorized,
				"Error getting the collection id")
		}

		if !uo.AccessAll || uo.Atype == model.UOTypeAdmin {
			_, err = core.ucs.FindByCollectionUser(cUuid, user.Uuid)
			if err != nil {
				return c.String(http.StatusUnauthorized,
					"The current user isn't a manager for this collection")
			}
		}

		c.Set(userKey, user)
		c.Set(deviceKey, device)
		c.Set(userOrganizationKey, uo)

		return next(c)
	}
}

// Not check collection id
func (core Core) RequireManagerLooseAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, device, err := core.baseAuth(c)
		if err != nil {
			return err
		}

		uo, err := core.orgAuth(c, user)
		if err != nil {
			return err
		}

		if uo.Atype > model.UOTypeAdmin && uo.Atype != model.UOTypeManager {
			return c.String(http.StatusUnauthorized,
				"You need to be a Manager, Admin or Owner to call this endpoint")
		}

		c.Set(userKey, user)
		c.Set(deviceKey, device)
		c.Set(userOrganizationKey, uo)

		return next(c)
	}
}

func (core Core) RequireOwnerAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		user, device, err := core.baseAuth(c)
		if err != nil {
			return err
		}

		uo, err := core.orgAuth(c, user)
		if err != nil {
			return err
		}

		if uo.Atype != model.UOTypeOwner {
			return c.String(http.StatusUnauthorized,
				"You need to be Owner to call this endpoint")
		}

		c.Set(userKey, user)
		c.Set(deviceKey, device)

		return next(c)
	}
}

func (core Core) baseAuth(c echo.Context) (*model.User, *model.Device, error) {

	rawToken := c.Request().Header.Get("Authorization")
	ts := strings.TrimPrefix(rawToken, "Bearer ")
	if ts == "" {
		return nil, nil, errors.New("missing authorization header")
	}

	claims := new(LoginJwtClaims)
	if err := core.DecodeToken(ts, claims); err != nil {
		core.logger.Debug().Err(err).Msg("")
		return nil, nil, err
	}

	device, err := core.devices.FindByUuid(claims.Device)
	if err != nil {
		core.logger.Debug().Err(err).Str("device uuid", claims.Device).Msg("")
		return nil, nil, c.String(http.StatusUnauthorized, "Invalid device id")
	}

	user, err := core.users.FindByUuid(claims.Subject)
	if err != nil {
		core.logger.Debug().Err(err).Str("user uuid", claims.Subject).Msg("")
		return nil, nil, c.String(http.StatusUnauthorized, "Invalid user id")
	}

	if user.SecurityStamp != claims.Sstamp {

		ssException := new(model.UserStampException)

		if err := json.Unmarshal([]byte(user.SecurityStamp), ssException); err != nil {
			core.logger.Debug().Err(err).Str("security stamp", user.SecurityStamp).Msg("")
			return nil, nil, c.String(http.StatusUnauthorized, "Invalid security stamp")
		}

		if time.Now().After(ssException.Expire) {
			empty := ""
			uu := &model.UpdateUser{
				Uuid:           user.Uuid,
				StampException: &empty,
			}

			if err := core.users.Update(uu); err != nil {
				core.logger.Debug().Err(err).Str("user uuid", user.Uuid).Msg("")
				return nil, nil, err
			}

			return nil, nil, c.String(http.StatusUnauthorized, "Stamp exception is expired")
		}
		// TODO: check security stamp
	}

	return user, device, nil
}

func (core Core) orgAuth(c echo.Context, user *model.User) (*model.UserOrganization, error) {
	orgUuid := getOrgUuid(c)

	uo, err := core.uos.FindByUserAndOrg(user.Uuid, orgUuid)
	if err != nil {
		return nil, c.String(http.StatusUnauthorized, "The current user isn't member of the organization")
	}

	if uo.Status != model.UOStatusConfirmed {
		return nil, c.String(http.StatusUnauthorized, "The current user isn't confirmed member of the organization")
	}

	return uo, nil
}

// Get org uuid from path
// /organizations/:uuid
// ?organizationId=:uuid
func getOrgUuid(c echo.Context) string {
	orgUuid := c.ParamValues()[0]
	if _, err := uuid.Parse(orgUuid); err == nil {
		return orgUuid
	}

	orgUuid = c.QueryParam("organizationId")
	if _, err := uuid.Parse(orgUuid); err == nil {
		return orgUuid
	}

	return ""
}

// Get collection uuid from path
// /organizations/:uuid/collections/:uuid
// ?collectionId=:uuid
func getCollectionUuid(c echo.Context) string {
	cUuid := c.ParamValues()[1]
	if _, err := uuid.Parse(cUuid); err == nil {
		return cUuid
	}

	cUuid = c.QueryParam("collectionId")
	if _, err := uuid.Parse(cUuid); err == nil {
		return cUuid
	}

	return ""
}
