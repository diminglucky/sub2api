package handler

import (
	"encoding/base64"
	"errors"
	"mime"
	"net/http"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

const maxImageStudioGalleryBytes = 32 << 20

type ImageStudioGalleryHandler struct {
	gallery *service.ImageStudioGalleryService
}

type saveImageStudioGalleryRequest struct {
	ImageDataURL string `json:"image_data_url"`
	ImageURL     string `json:"image_url"`
	Prompt       string `json:"prompt"`
	Model        string `json:"model"`
	Size         string `json:"size"`
	Format       string `json:"format"`
}

func NewImageStudioGalleryHandler(gallery *service.ImageStudioGalleryService) *ImageStudioGalleryHandler {
	return &ImageStudioGalleryHandler{gallery: gallery}
}

func (h *ImageStudioGalleryHandler) Save(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	var request saveImageStudioGalleryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.BadRequest(c, "Invalid gallery request")
		return
	}

	ctx := c.Request.Context()
	dataURL := strings.TrimSpace(request.ImageDataURL)
	imageURL := strings.TrimSpace(request.ImageURL)

	var entry *service.ImageStudioGalleryEntry
	var err error
	switch {
	case dataURL != "":
		var data []byte
		var contentType string
		data, contentType, err = decodeImageStudioDataURL(dataURL)
		if err != nil {
			response.BadRequest(c, err.Error())
			return
		}
		entry, err = h.gallery.Save(
			ctx, subject.UserID, request.Prompt, request.Model, request.Size, request.Format, contentType, data,
		)
	case imageURL != "":
		entry, err = h.gallery.SaveFromURL(
			ctx, subject.UserID, request.Prompt, request.Model, request.Size, request.Format, imageURL,
		)
	default:
		response.BadRequest(c, "image_data_url or image_url is required")
		return
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, entry)
}

func (h *ImageStudioGalleryHandler) List(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	items, err := h.gallery.List(c.Request.Context(), subject.UserID, 100)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}

func decodeImageStudioDataURL(raw string) ([]byte, string, error) {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(strings.ToLower(raw), "data:") {
		return nil, "", errors.New("image_data_url must be a data URL")
	}
	header, payload, ok := strings.Cut(raw[len("data:"):], ",")
	if !ok {
		return nil, "", errors.New("invalid image data URL")
	}
	parts := strings.Split(header, ";")
	if len(parts) < 2 || !strings.EqualFold(strings.TrimSpace(parts[len(parts)-1]), "base64") {
		return nil, "", errors.New("image data URL must be base64 encoded")
	}
	contentType, _, err := mime.ParseMediaType(strings.Join(parts[:len(parts)-1], ";"))
	if err != nil || !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return nil, "", errors.New("image data URL must contain an image")
	}
	if len(payload) > base64.StdEncoding.EncodedLen(maxImageStudioGalleryBytes) {
		return nil, "", errors.New("image is too large")
	}
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return nil, "", errors.New("invalid base64 image")
	}
	if len(data) == 0 || len(data) > maxImageStudioGalleryBytes {
		return nil, "", errors.New("image is too large or empty")
	}
	detected := strings.TrimSpace(strings.Split(http.DetectContentType(data), ";")[0])
	if strings.HasPrefix(detected, "image/") {
		contentType = detected
	}
	return data, contentType, nil
}
