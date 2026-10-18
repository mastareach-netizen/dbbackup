package adapters

import (
	"fmt"

	"dbbackup/src/internal/adapters/common"
)

type (
	BackupMode       = common.BackupMode
	ConnectionConfig = common.ConnectionConfig
	RestoreOptions   = common.RestoreOptions
	Adapter          = common.Adapter
)

const ModeFull = common.ModeFull

func ValidateMode(adapter Adapter, mode BackupMode) error {
	for _, m := range adapter.SupportedBackupModes() {
		if m == mode {
			return nil
		}
	}
	return fmt.Errorf("%s does not support backup mode %q (supported: %v)", adapter.Name(), mode, adapter.SupportedBackupModes())
}
