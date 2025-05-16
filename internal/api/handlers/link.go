package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

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
		switch r.Method {
		case http.MethodPost:
			lh.CreateLinkHandler().ServeHTTP(w, r)
		case http.MethodGet:
			if r.URL.Query().Has("id") {
				lh.GetLinkHandler().ServeHTTP(w, r)
			} else {
				lh.ListLinksHandler().ServeHTTP(w, r)
			}
		case http.MethodPut:
			lh.UpdateLinkHandler().ServeHTTP(w, r)
		case http.MethodDelete:
			lh.DeleteLinkHandler().ServeHTTP(w, r)
		default:
			utils.WriteJSONError(w, http.StatusMethodNotAllowed, "Method Not Allowed")
		}
	}
}

func (lh *LinkHandler) CreateLinkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodPost) {
			return
		}
		if _, ok := verifySubFromContext(w, r); !ok {
			return
		}
		userID, ok := extractUserIDFromRequest(w, r)
		if !ok {
			return
		}
		if !checkUserExists(r.Context(), lh.UserService, w, userID) {
			return
		}

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
}

func (lh *LinkHandler) ListLinksHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodGet) {
			return
		}
		if _, ok := verifySubFromContext(w, r); !ok {
			return
		}
		userID, ok := extractUserIDFromRequest(w, r)
		if !ok {
			return
		}
		if !checkUserExists(r.Context(), lh.UserService, w, userID) {
			return
		}
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
}

func (lh *LinkHandler) DeleteLinkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodDelete) {
			return
		}
		if _, ok := verifySubFromContext(w, r); !ok {
			return
		}
		userID, ok := extractUserIDFromRequest(w, r)
		if !ok {
			return
		}
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
}

func (lh *LinkHandler) UpdateLinkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.WriteJSONError(w, http.StatusNotImplemented, "Update link not implemented yet")
	}
}
func (lh *LinkHandler) GetLinkHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodGet) {
			return
		}
		if _, ok := verifySubFromContext(w, r); !ok {
			return
		}
		userID, ok := extractUserIDFromRequest(w, r)
		if !ok {
			return
		}
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
}

func (lh *LinkHandler) RedirectHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		shortCode := strings.TrimPrefix(r.URL.Path, "/r/")
		if shortCode == "" {
			utils.WriteJSONError(w, http.StatusBadRequest, "Missing short code")
			return
		}
		go func() {
			ip := r.RemoteAddr
			ref := r.Referer()
			ua := r.UserAgent()
			_ = lh.LinkService.TrackClick(r.Context(), shortCode, ip, ref, ua)
		}()
		destination, _, err := lh.LinkService.ResolveLink(r.Context(), shortCode)
		if err == link.ErrLinkNotFound {
			utils.WriteJSONError(w, http.StatusNotFound, "Link not found")
			return
		} else if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "Could not resolve link")
			return
		}
		http.Redirect(w, r, destination, http.StatusFound)
	}
}

func (lh *LinkHandler) ClickCountHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !validateMethod(w, r, http.MethodGet) {
			return
		}
		shortCode := r.URL.Query().Get("short_code")
		if shortCode == "" {
			utils.WriteJSONError(w, http.StatusBadRequest, "Missing short_code")
			return
		}
		count, err := lh.LinkService.GetClickCount(r.Context(), shortCode)
		if err != nil {
			utils.WriteJSONError(w, http.StatusInternalServerError, "Failed to get click count")
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]int{
			"click_count": count,
		})
	}
}
