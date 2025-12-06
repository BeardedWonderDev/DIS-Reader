package bridgeproto

// Public aliases for bridge proto messages/enums so callers outside the module
// can reference the types without importing an internal package path.

import internalproto "github.com/BeardedWonderDev/DIS-Reader/internal/bridge/proto"

type (
	Status             = internalproto.Status
	JobKind            = internalproto.JobKind
	JobRequest         = internalproto.JobRequest
	JobResult          = internalproto.JobResult
	Row                = internalproto.Row
	AgentConfig        = internalproto.AgentConfig
	AgentRuntimeConfig = internalproto.AgentRuntimeConfig
	AgentRuntimeStatus = internalproto.AgentRuntimeStatus
	AgentConfigStatus  = internalproto.AgentConfigStatus
	LokiConfig         = internalproto.LokiConfig
)

// Re-export enums so consumers can reference them without the internal path.
const (
	Status_STATUS_UNSPECIFIED = internalproto.Status_STATUS_UNSPECIFIED
	Status_STATUS_OK          = internalproto.Status_STATUS_OK
	Status_STATUS_ERROR       = internalproto.Status_STATUS_ERROR
	Status_STATUS_DONE        = internalproto.Status_STATUS_DONE
)

const (
	JobKind_JOB_KIND_UNSPECIFIED   = internalproto.JobKind_JOB_KIND_UNSPECIFIED
	JobKind_JOB_KIND_QUERY         = internalproto.JobKind_JOB_KIND_QUERY
	JobKind_JOB_KIND_PING_SERVICE  = internalproto.JobKind_JOB_KIND_PING_SERVICE
	JobKind_JOB_KIND_PING_DATABASE = internalproto.JobKind_JOB_KIND_PING_DATABASE
	JobKind_JOB_KIND_CONNECT       = internalproto.JobKind_JOB_KIND_CONNECT
	JobKind_JOB_KIND_DISCONNECT    = internalproto.JobKind_JOB_KIND_DISCONNECT
	JobKind_JOB_KIND_START_JDBC    = internalproto.JobKind_JOB_KIND_START_JDBC
	JobKind_JOB_KIND_STOP_JDBC     = internalproto.JobKind_JOB_KIND_STOP_JDBC
	JobKind_JOB_KIND_READ_CONFIG   = internalproto.JobKind_JOB_KIND_READ_CONFIG
)
