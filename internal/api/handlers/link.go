package handlers

import (
	"encoding/json"
	"net/http"

	lru "github.com/hashicorp/golang-lru"
	"redo.ai/internal/model"
	"redo.ai/internal/pkg/platform"
	"redo.ai/internal/service/link"
	"redo.ai/internal/service/user"
	"redo.ai/internal/utils"
	"redo.ai/logger"
)

type LinkHandler struct {
	LinkService link.LinkService
	UserService user.UserService
	Platform    platform.PlatformDetector
	Cache       *lru.Cache
}

func NewLinkHandler(userService user.UserService, linkService link.LinkService, cache *lru.Cache) *LinkHandler {
	return &LinkHandler{
		LinkService: linkService,
		UserService: userService,
		Cache:       cache,
	}
}

func (lh *LinkHandler) LinksRouter() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, ok := verifySubFromContext(w, r); !ok {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
		userID, ok := extractUserIDFromRequest(w, r)
		if !ok {
			utils.WriteJSONError(w, http.StatusUnauthorized, "Invalid user")
			return
		}

		if !checkUserExists(r.Context(), lh.UserService, w, userID) {
			return
		}

		switch r.Method {
		case http.MethodPost:
			lh.CreateLinkHandler(w, r, userID)
		case http.MethodGet:
			if r.URL.Query().Has("id") {
				lh.GetLinkHandler(w, r, userID)
			} else {
				lh.ListLinksHandler(w, r, userID)
			}
		case http.MethodPut:
			lh.UpdateLinkHandler(w, r, userID)
		case http.MethodDelete:
			lh.DeleteLinkHandler(w, r, userID)
		default:
			utils.WriteJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	}
}

func (lh *LinkHandler) CreateLinkHandler(w http.ResponseWriter, r *http.Request, userID string) {
	var req model.CreateLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}
	if (req.Slug != "" && !utils.IsValidSlug(req.Slug)) || !utils.IsValidURL(req.Destination) {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid format")
		return
	}

	lk, err := lh.LinkService.CreateLink(r.Context(), userID, req)
	if err != nil {
		if err == link.ErrSlugAlreadyExists {
			utils.WriteJSONError(w, http.StatusConflict, "Slug already exists")
			return
		}
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to create link")
		return
	}

	if cached, ok := lh.Cache.Get(userID); ok {
		if links, ok := cached.([]model.Link); ok {
			lh.Cache.Add(userID, append(links, lk))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

}

func (lh *LinkHandler) ListLinksHandler(w http.ResponseWriter, r *http.Request, userID string) {
	if cached, ok := lh.Cache.Get(userID); ok {
		if links, ok := cached.([]model.Link); ok {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(links)
			return
		}
	}
	links, err := lh.LinkService.ListLinks(r.Context(), userID)
	if err != nil {
		logger.Error("ListLinksHandler: failed to fetch links: %v", err)
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to fetch links")
		return
	}
	lh.Cache.Add(userID, links)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(links)

}

func (lh *LinkHandler) DeleteLinkHandler(w http.ResponseWriter, r *http.Request, userID string) {
	linkID := r.URL.Query().Get("linkId")
	if linkID == "" || !IsValidUUID(linkID) {
		logger.Error("DeleteLinkHandler: Invalid or missing link ID")
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid or missing link ID")
		return
	}
	if err := lh.LinkService.DeleteLink(r.Context(), userID, linkID); err != nil {
		if err == link.ErrLinkNotFound {
			logger.Error("DeleteLinkHandler: link not found or unauthorized: %v", err)
			utils.WriteJSONError(w, http.StatusNotFound, "Link not found or unauthorized")
		} else {
			logger.Error("DeleteLinkHandler: failed to delete links: %v", err)
			utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to delete link")
		}
		return
	}
	logger.Info("link deleted")
	w.WriteHeader(http.StatusNoContent)
}

func (lh *LinkHandler) UpdateLinkHandler(w http.ResponseWriter, r *http.Request, userID string) {
	utils.WriteJSONError(w, http.StatusNotImplemented, "Update link not implemented yet")
}

func (lh *LinkHandler) GetLinkHandler(w http.ResponseWriter, r *http.Request, userID string) {
	linkID := r.URL.Query().Get("id")
	if linkID == "" || !IsValidUUID(linkID) {
		utils.WriteJSONError(w, http.StatusBadRequest, "Invalid or missing link ID")
		return
	}
	links, err := lh.LinkService.ListLinks(r.Context(), userID)
	if err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to retrieve links")
		return
	}
	for _, link := range links {
		if link.LinkID == linkID {
			_ = json.NewEncoder(w).Encode(link)
			return
		}
	}
	utils.WriteJSONError(w, http.StatusNotFound, "Link not found")
}
