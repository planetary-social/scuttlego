package commands

type DisconnectAllHandler struct {
	peerManager PeerManager
}

func NewDisconnectAllHandler(peerManager PeerManager) *DisconnectAllHandler {
	return &DisconnectAllHandler{peerManager: peerManager}
}

func (h *DisconnectAllHandler) Handle() error {
	return h.peerManager.DisconnectAll()
}
