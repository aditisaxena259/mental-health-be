package helpers

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type CloudinaryClient struct {
	cld *cloudinary.Cloudinary
}

var cloudClient *CloudinaryClient

// InitCloudinary initializes a singleton Cloudinary client using CLOUDINARY_URL or discrete env vars
// Supported env:
// - CLOUDINARY_URL (cloudinary://<api_key>:<api_secret>@<cloud_name>)
// - or CLOUDINARY_CLOUD_NAME, CLOUDINARY_API_KEY, CLOUDINARY_API_SECRET
func InitCloudinary() (*CloudinaryClient, error) {
	if cloudClient != nil {
		return cloudClient, nil
	}

	var cld *cloudinary.Cloudinary
	var err error

	if raw := os.Getenv("CLOUDINARY_URL"); raw != "" {
		cld, err = cloudinary.NewFromURL(raw)
		if err != nil {
			return nil, fmt.Errorf("cloudinary init from URL failed: %w", err)
		}
	} else {
		cloudName := os.Getenv("CLOUDINARY_CLOUD_NAME")
		apiKey := os.Getenv("CLOUDINARY_API_KEY")
		apiSecret := os.Getenv("CLOUDINARY_API_SECRET")
		if cloudName == "" || apiKey == "" || apiSecret == "" {
			return nil, errors.New("cloudinary credentials not configured: set CLOUDINARY_URL or CLOUDINARY_CLOUD_NAME/API_KEY/API_SECRET")
		}
		// Construct URL manually for sdk
		u := &url.URL{Scheme: "cloudinary", Host: fmt.Sprintf("%s:%s@%s", apiKey, apiSecret, cloudName)}
		cld, err = cloudinary.NewFromURL(u.String())
		if err != nil {
			return nil, fmt.Errorf("cloudinary init failed: %w", err)
		}
	}

	cloudClient = &CloudinaryClient{cld: cld}
	return cloudClient, nil
}

type UploadResult struct {
	SecureURL string
	PublicID  string
	Width     int
	Height    int
	Bytes     int64
	Format    string
}

// UploadJPEG uploads a local file path as image to Cloudinary under folder and returns public id and URL
func (c *CloudinaryClient) UploadJPEG(localPath, folder, publicIDHint string) (*UploadResult, error) {
	if c == nil || c.cld == nil {
		return nil, errors.New("cloudinary not initialized")
	}
	params := uploader.UploadParams{
		Folder:       folder,
		PublicID:     publicIDHint,
		ResourceType: "image",
		Overwrite:    api.Bool(true),
		Invalidate:   api.Bool(true),
	}
	res, err := c.cld.Upload.Upload(ctx(), localPath, params)
	if err != nil {
		return nil, err
	}
	return &UploadResult{
		SecureURL: res.SecureURL,
		PublicID:  res.PublicID,
		Width:     res.Width,
		Height:    res.Height,
		Bytes:     int64(res.Bytes),
		Format:    res.Format,
	}, nil
}

// Destroy removes an asset by public id
func (c *CloudinaryClient) Destroy(publicID string) error {
	if c == nil || c.cld == nil {
		return errors.New("cloudinary not initialized")
	}
	_, err := c.cld.Upload.Destroy(ctx(), uploader.DestroyParams{PublicID: publicID, Invalidate: api.Bool(true), ResourceType: "image"})
	return err
}

// context helper (lazy)
func ctx() context.Context {
	return context.Background()
}
