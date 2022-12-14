// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/rs/zerolog"
	"github.com/togls/gowarden/auth"
	"github.com/togls/gowarden/config"
	"github.com/togls/gowarden/handler"
	"github.com/togls/gowarden/handler/middleware"
	"github.com/togls/gowarden/store/raw"
)

import (
	_ "github.com/go-sql-driver/mysql"
)

// Injectors from wire.go:

func createApp(configFile string, log *zerolog.Logger) (*Apllication, error) {
	core, err := config.New(configFile, log)
	if err != nil {
		return nil, err
	}
	db, err := OpenDB(core)
	if err != nil {
		return nil, err
	}
	user := raw.NewUserStore(db)
	device := raw.NewDeviceStore(db)
	userOrganization := raw.NewUserOrganizationStore(db)
	send := raw.NewSendStore(db)
	emergencyAccess := raw.NewEmergencyAccessStore(db)
	cipher := raw.NewCipherStore(db)
	favorite := raw.NewFavoriteStore(db)
	folder := raw.NewFolderStore(db)
	twoFactor := raw.NewTwoFactorStore(db)
	twoFactorIncomplete := raw.NewTwoFactorIncompleteStore(db)
	invitation := raw.NewInvitationStore(db)
	userCollection := raw.NewUserCollectionStore(db)
	authCore := auth.New(core, device, user, userOrganization, userCollection)
	accountHandler := handler.NewAccountHandler(user, device, userOrganization, send, emergencyAccess, cipher, favorite, folder, twoFactor, twoFactorIncomplete, invitation, authCore, authCore)
	globalDomains, err := config.LoadGlobalDomain()
	if err != nil {
		return nil, err
	}
	attachment := raw.NewAttachmentStore(db)
	collection := raw.NewCollectionStore(db)
	orgPolicy := raw.NewOrgPolicyStore(db)
	cipherHandler := handler.NewCipherHandler(log, globalDomains, authCore, attachment, cipher, collection, favorite, folder, orgPolicy, send, userCollection, user, userOrganization)
	folderHandler := handler.NewFolderHandler(core, folder, authCore)
	organization := raw.NewOrganizationStore(db)
	organizationHandler := handler.NewOrganizationHandler(user, cipher, organization, collection, orgPolicy, userOrganization, userCollection, invitation, authCore, core)
	iconHandler := handler.NewIconHandler()
	identityHandler := handler.NewIdentityHandler(log, authCore)
	appHeader := middleware.NewAppHeader(core)
	middlewareRecover := middleware.NewRecover(log)
	logger := middleware.NewLogger(log)
	muxOptions := &handler.MuxOptions{
		Cfg:          core,
		Account:      accountHandler,
		Cipher:       cipherHandler,
		Folder:       folderHandler,
		Organization: organizationHandler,
		Icon:         iconHandler,
		Identity:     identityHandler,
		AppHeader:    appHeader,
		Recover:      middlewareRecover,
		LoggerMW:     logger,
	}
	httpHandler := handler.NewMux(muxOptions)
	mainOptions := options{
		cfg:     core,
		handler: httpHandler,
		logger:  log,
		db:      db,
	}
	apllication := NewApplication(mainOptions)
	return apllication, nil
}
