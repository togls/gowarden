package raw

import (
	"github.com/google/wire"
)

var WireSet = wire.NewSet(
	NewAttachmentStore,
	NewCipherStore,
	NewCollectionStore,
	NewDeviceStore,
	NewEmergencyAccessStore,
	NewFavoriteStore,
	NewFolderStore,
	NewInvitationStore,
	NewOrgPolicyStore,
	NewOrganizationStore,
	NewSendStore,
	NewTwoFactorStore,
	NewTwoFactorIncompleteStore,
	NewUserCollectionStore,
	NewUserOrganizationStore,
	NewUserStore,
)
