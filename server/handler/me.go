package handler

import (
	"net/http"

	"github.com/pdrhlik/edemos/server/identity"
)

func (h *Handler) Me() AppHandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		ident := identity.GetUserFromContext(r.Context())
		if ident == nil {
			return writeError(w, http.StatusUnauthorized, "unauthorized")
		}

		u, err := h.Store.GetUserByID(r.Context(), ident.ID)
		if err != nil {
			return err
		}
		if u == nil {
			return writeError(w, http.StatusUnauthorized, "user not found")
		}

		return writeJSON(w, http.StatusOK, u)
	}
}
