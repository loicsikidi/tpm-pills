package tpmutil

import "path"

const SWTPM_ROOT_STATE = ".swtpm"

var SWTPM_STATE = path.Join(SWTPM_ROOT_STATE, "state")
