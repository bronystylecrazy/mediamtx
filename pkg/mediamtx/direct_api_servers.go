package mediamtx

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/bluenviron/mediamtx/pkg/mediamtx/defs"
)

// =============================================================================
// RTMP SERVER MANAGEMENT
// =============================================================================

// GetRTMPConnections returns a list of RTMP connections with pagination
func (api *DirectAPI) GetRTMPConnections(pagination *PaginationParams) (*defs.APIRTMPConnList, error) {
	if api.core.RtmpServer == nil {
		return nil, fmt.Errorf("RTMP server not available")
	}
	
	data, err := api.core.RtmpServer.APIConnsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get RTMP connections: %v", err)
	}
	
	data.ItemCount = len(data.Items)
	
	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}
	
	return data, nil
}

// GetRTMPConnection returns information about a specific RTMP connection
func (api *DirectAPI) GetRTMPConnection(id string) (*defs.APIRTMPConn, error) {
	if api.core.RtmpServer == nil {
		return nil, fmt.Errorf("RTMP server not available")
	}
	
	connUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %v", err)
	}
	
	data, err := api.core.RtmpServer.APIConnsGet(connUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get RTMP connection: %v", err)
	}
	
	return data, nil
}

// KickRTMPConnection kicks (disconnects) an RTMP connection
func (api *DirectAPI) KickRTMPConnection(id string) error {
	if api.core.RtmpServer == nil {
		return fmt.Errorf("RTMP server not available")
	}
	
	connUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %v", err)
	}
	
	err = api.core.RtmpServer.APIConnsKick(connUUID)
	if err != nil {
		return fmt.Errorf("failed to kick RTMP connection: %v", err)
	}
	
	return nil
}

// GetRTMPSConnections returns a list of RTMPS connections with pagination
func (api *DirectAPI) GetRTMPSConnections(pagination *PaginationParams) (*defs.APIRTMPConnList, error) {
	if api.core.RtmpsServer == nil {
		return nil, fmt.Errorf("RTMPS server not available")
	}
	
	data, err := api.core.RtmpsServer.APIConnsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get RTMPS connections: %v", err)
	}
	
	data.ItemCount = len(data.Items)
	
	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}
	
	return data, nil
}

// GetRTMPSConnection returns information about a specific RTMPS connection
func (api *DirectAPI) GetRTMPSConnection(id string) (*defs.APIRTMPConn, error) {
	if api.core.RtmpsServer == nil {
		return nil, fmt.Errorf("RTMPS server not available")
	}
	
	connUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %v", err)
	}
	
	data, err := api.core.RtmpsServer.APIConnsGet(connUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get RTMPS connection: %v", err)
	}
	
	return data, nil
}

// KickRTMPSConnection kicks (disconnects) an RTMPS connection
func (api *DirectAPI) KickRTMPSConnection(id string) error {
	if api.core.RtmpsServer == nil {
		return fmt.Errorf("RTMPS server not available")
	}
	
	connUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %v", err)
	}
	
	err = api.core.RtmpsServer.APIConnsKick(connUUID)
	if err != nil {
		return fmt.Errorf("failed to kick RTMPS connection: %v", err)
	}
	
	return nil
}

// =============================================================================
// HLS SERVER MANAGEMENT
// =============================================================================

// GetHLSMuxers returns a list of HLS muxers with pagination
func (api *DirectAPI) GetHLSMuxers(pagination *PaginationParams) (*defs.APIHLSMuxerList, error) {
	if api.core.HlsServer == nil {
		return nil, fmt.Errorf("HLS server not available")
	}
	
	data, err := api.core.HlsServer.APIMuxersList()
	if err != nil {
		return nil, fmt.Errorf("failed to get HLS muxers: %v", err)
	}
	
	data.ItemCount = len(data.Items)
	
	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}
	
	return data, nil
}

// GetHLSMuxer returns information about a specific HLS muxer
func (api *DirectAPI) GetHLSMuxer(name string) (*defs.APIHLSMuxer, error) {
	if api.core.HlsServer == nil {
		return nil, fmt.Errorf("HLS server not available")
	}
	
	data, err := api.core.HlsServer.APIMuxersGet(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get HLS muxer: %v", err)
	}
	
	return data, nil
}

// =============================================================================
// WEBRTC SERVER MANAGEMENT
// =============================================================================

// GetWebRTCSessions returns a list of WebRTC sessions with pagination
func (api *DirectAPI) GetWebRTCSessions(pagination *PaginationParams) (*defs.APIWebRTCSessionList, error) {
	if api.core.WebRTCServer == nil {
		return nil, fmt.Errorf("WebRTC server not available")
	}
	
	data, err := api.core.WebRTCServer.APISessionsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get WebRTC sessions: %v", err)
	}
	
	data.ItemCount = len(data.Items)
	
	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}
	
	return data, nil
}

// GetWebRTCSession returns information about a specific WebRTC session
func (api *DirectAPI) GetWebRTCSession(id string) (*defs.APIWebRTCSession, error) {
	if api.core.WebRTCServer == nil {
		return nil, fmt.Errorf("WebRTC server not available")
	}
	
	sessionUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %v", err)
	}
	
	data, err := api.core.WebRTCServer.APISessionsGet(sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get WebRTC session: %v", err)
	}
	
	return data, nil
}

// KickWebRTCSession kicks (disconnects) a WebRTC session
func (api *DirectAPI) KickWebRTCSession(id string) error {
	if api.core.WebRTCServer == nil {
		return fmt.Errorf("WebRTC server not available")
	}
	
	sessionUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid session ID: %v", err)
	}
	
	err = api.core.WebRTCServer.APISessionsKick(sessionUUID)
	if err != nil {
		return fmt.Errorf("failed to kick WebRTC session: %v", err)
	}
	
	return nil
}

// =============================================================================
// SRT SERVER MANAGEMENT
// =============================================================================

// GetSRTConnections returns a list of SRT connections with pagination
func (api *DirectAPI) GetSRTConnections(pagination *PaginationParams) (*defs.APISRTConnList, error) {
	if api.core.SrtServer == nil {
		return nil, fmt.Errorf("SRT server not available")
	}
	
	data, err := api.core.SrtServer.APIConnsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get SRT connections: %v", err)
	}
	
	data.ItemCount = len(data.Items)
	
	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}
	
	return data, nil
}

// GetSRTConnection returns information about a specific SRT connection
func (api *DirectAPI) GetSRTConnection(id string) (*defs.APISRTConn, error) {
	if api.core.SrtServer == nil {
		return nil, fmt.Errorf("SRT server not available")
	}
	
	connUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %v", err)
	}
	
	data, err := api.core.SrtServer.APIConnsGet(connUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get SRT connection: %v", err)
	}
	
	return data, nil
}

// KickSRTConnection kicks (disconnects) an SRT connection
func (api *DirectAPI) KickSRTConnection(id string) error {
	if api.core.SrtServer == nil {
		return fmt.Errorf("SRT server not available")
	}
	
	connUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid connection ID: %v", err)
	}
	
	err = api.core.SrtServer.APIConnsKick(connUUID)
	if err != nil {
		return fmt.Errorf("failed to kick SRT connection: %v", err)
	}
	
	return nil
}

// =============================================================================
// RTSPS (RTSP over TLS) SERVER MANAGEMENT
// =============================================================================

// GetRTSPSConnections returns a list of RTSPS connections with pagination
func (api *DirectAPI) GetRTSPSConnections(pagination *PaginationParams) (*defs.APIRTSPConnsList, error) {
	if api.core.RtspsServer == nil {
		return nil, fmt.Errorf("RTSPS server not available")
	}
	
	data, err := api.core.RtspsServer.APIConnsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get RTSPS connections: %v", err)
	}
	
	data.ItemCount = len(data.Items)
	
	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}
	
	return data, nil
}

// GetRTSPSConnection returns information about a specific RTSPS connection
func (api *DirectAPI) GetRTSPSConnection(id string) (*defs.APIRTSPConn, error) {
	if api.core.RtspsServer == nil {
		return nil, fmt.Errorf("RTSPS server not available")
	}
	
	connUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid connection ID: %v", err)
	}
	
	data, err := api.core.RtspsServer.APIConnsGet(connUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get RTSPS connection: %v", err)
	}
	
	return data, nil
}

// GetRTSPSSessions returns a list of RTSPS sessions with pagination
func (api *DirectAPI) GetRTSPSSessions(pagination *PaginationParams) (*defs.APIRTSPSessionList, error) {
	if api.core.RtspsServer == nil {
		return nil, fmt.Errorf("RTSPS server not available")
	}
	
	data, err := api.core.RtspsServer.APISessionsList()
	if err != nil {
		return nil, fmt.Errorf("failed to get RTSPS sessions: %v", err)
	}
	
	data.ItemCount = len(data.Items)
	
	if pagination != nil {
		pageCount := api.paginateSlice(&data.Items, pagination.ItemsPerPage, pagination.Page)
		data.PageCount = pageCount
	} else {
		data.PageCount = 1
	}
	
	return data, nil
}

// GetRTSPSSession returns information about a specific RTSPS session
func (api *DirectAPI) GetRTSPSSession(id string) (*defs.APIRTSPSession, error) {
	if api.core.RtspsServer == nil {
		return nil, fmt.Errorf("RTSPS server not available")
	}
	
	sessionUUID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid session ID: %v", err)
	}
	
	data, err := api.core.RtspsServer.APISessionsGet(sessionUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get RTSPS session: %v", err)
	}
	
	return data, nil
}

// KickRTSPSSession kicks (disconnects) an RTSPS session
func (api *DirectAPI) KickRTSPSSession(id string) error {
	if api.core.RtspsServer == nil {
		return fmt.Errorf("RTSPS server not available")
	}
	
	sessionUUID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid session ID: %v", err)
	}
	
	err = api.core.RtspsServer.APISessionsKick(sessionUUID)
	if err != nil {
		return fmt.Errorf("failed to kick RTSPS session: %v", err)
	}
	
	return nil
}