package constants

import "errors"

var ErrPendingDatamigrateMigration = errors.New("tbl_datamigrate_applied table does not exist; run mage dbup to apply pending migrations")

const (
	MsgFailedCheckDatamigrateTable         = "failed to check datamigrate table"
	MsgFailedCheckGooseTable               = "failed to check goose migration table"
	MsgFailedReadGooseVersion              = "failed to read goose version"
	MsgFailedListAppliedDatamigrateScripts = "failed to list applied datamigrate scripts"
	MsgFailedRecordDatamigrateScript       = "failed to record datamigrate script "
	MsgFailedMarkDatamigrateScript         = "failed to mark datamigrate script "
	MsgUnknownDatamigrateScript            = "unknown datamigrate script: "
)
