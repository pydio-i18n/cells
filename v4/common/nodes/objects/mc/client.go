/*
 * Copyright (c) 2019-2021. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package mc

import (
	"context"
	"github.com/minio/minio-go/v7/pkg/notification"
	"io"
	"strings"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/pydio/cells/v4/common/nodes"
	"github.com/pydio/cells/v4/common/nodes/models"
	"github.com/pydio/cells/v4/x/configx"
)

// Client wraps a minio.Core client in the nodes.StorageClient interface
type Client struct {
	mc *minio.Core
}

func init() {
	nodes.RegisterStorageClient("mc", func(cfg configx.Values) (nodes.StorageClient, error) {
		ep := cfg.Val("endpoint").String()
		key := cfg.Val("key").String()
		secret := cfg.Val("secret").String()
		secure := cfg.Val("secure").Bool()
		return New(ep, key, secret, secure)
	})
}

// New creates a new minio.Core with the most standard options
func New(endpoint, accessKey, secretKey string, secure bool, customRegion ...string) (*Client, error) {
	options := &minio.Options{
		Creds:  credentials.NewStaticV2(accessKey, secretKey, ""),
		Secure: secure,
	}
	if len(customRegion) > 0 {
		options.Region = customRegion[0]
	}
	c, err := minio.NewCore(endpoint, options)
	if err != nil {
		return nil, err
	}
	return &Client{
		mc: c,
	}, nil
}

func (c *Client) ListBucketsWithContext(ctx context.Context) ([]models.BucketInfo, error) {
	bb, e := c.mc.ListBuckets(ctx)
	if e != nil {
		return nil, e
	}
	buckets := make([]models.BucketInfo, len(bb))
	for i, b := range bb {
		buckets[i] = models.BucketInfo{
			Name:         b.Name,
			CreationDate: b.CreationDate,
		}
	}
	return buckets, nil
}

func (c *Client) MakeBucketWithContext(ctx context.Context, bucketName string, location string) (err error) {
	return c.mc.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
}

func (c *Client) RemoveBucketWithContext(ctx context.Context, bucketName string) error {
	return c.mc.RemoveBucket(ctx, bucketName)
}

func (c *Client) GetObject(bucketName, objectName string, opts models.ReadMeta) (io.ReadCloser, models.ObjectInfo, error) {
	getOpts := readMetaToMinioOpts(opts)
	rc, oi, _, e := c.mc.GetObject(context.Background(), bucketName, objectName, getOpts)
	if e != nil {
		return nil, models.ObjectInfo{}, e
	}
	return rc, minioInfoToModelsInfo(oi), nil
}

func (c *Client) GetObjectWithContext(ctx context.Context, bucketName, objectName string, opts models.ReadMeta) (io.ReadCloser, error) {
	rc, _, _, e := c.mc.GetObject(ctx, bucketName, objectName, readMetaToMinioOpts(opts))
	return rc, e
}

func (c *Client) StatObject(bucketName, objectName string, opts models.ReadMeta) (models.ObjectInfo, error) {
	getOpts := readMetaToMinioOpts(opts)
	oi, e := c.mc.StatObject(context.Background(), bucketName, objectName, getOpts)
	return minioInfoToModelsInfo(oi), e
}

func (c *Client) PutObject(bucket, object string, data io.Reader, size int64, md5Base64, sha256Hex string, metadata models.ReadMeta) (models.ObjectInfo, error) {
	oi, e := c.mc.PutObject(context.Background(), bucket, object, data, size, md5Base64, sha256Hex, putMetaToMinioOpts(models.PutMeta{UserMetadata: metadata}))
	return minioUploadInfoToModelsInfo(oi), e
}

func (c *Client) PutObjectWithContext(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64,
	opts models.PutMeta) (n int64, err error) {
	ui, e := c.mc.PutObject(ctx, bucketName, objectName, reader, objectSize, "", "", putMetaToMinioOpts(opts))
	if e != nil {
		return 0, e
	} else {
		return ui.Size, e
	}
}

func (c *Client) RemoveObjectWithContext(ctx context.Context, bucketName, objectName string) error {
	return c.mc.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

func (c *Client) ListObjectsWithContext(ctx context.Context, bucket, prefix, marker, delimiter string, maxKeys int) (result models.ListBucketResult, err error) {
	recursive := true
	if delimiter == "/" {
		recursive = false
	}
	ch := c.mc.Client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:     prefix,
		Recursive:  recursive,
		MaxKeys:    maxKeys,
		StartAfter: marker,
	})
	r := models.ListBucketResult{
		//IsTruncated:    res.IsTruncated,
		//EncodingType:   res.EncodingType,
		//NextMarker:     res.NextMarker,
		Delimiter: delimiter,
		Marker:    marker,
		MaxKeys:   int64(maxKeys),
		Name:      bucket,
		Prefix:    prefix,
	}

	for oi := range ch {
		if oi.Err != nil {
			return result, oi.Err
		}
		if strings.HasSuffix(oi.Key, "/") {
			r.CommonPrefixes = append(r.CommonPrefixes, models.CommonPrefix{Prefix: oi.Key})
		} else {
			r.Contents = append(r.Contents, minioInfoToModelsInfo(oi))
		}
	}

	/*
		r := models.ListBucketResult{
			CommonPrefixes: make([]models.CommonPrefix, len(res.CommonPrefixes)),
			Contents:       make([]models.ObjectInfo, len(res.Contents)),
			Delimiter:      res.Delimiter,
			EncodingType:   res.EncodingType,
			IsTruncated:    res.IsTruncated,
			Marker:         res.Marker,
			MaxKeys:        res.MaxKeys,
			Name:           res.Name,
			NextMarker:     res.NextMarker,
			Prefix:         res.Prefix,
		}
		for i, c := range res.Contents {
			r.Contents[i] = minioInfoToModelsInfo(c)
		}
		for i, cp := range res.CommonPrefixes {
			r.CommonPrefixes[i] = models.CommonPrefix{Prefix: cp.Prefix}
		}
	*/
	return r, nil
}

func (c *Client) NewMultipartUploadWithContext(ctx context.Context, bucket, object string, opts models.PutMeta) (uploadID string, err error) {
	return c.mc.NewMultipartUpload(ctx, bucket, object, putMetaToMinioOpts(opts))
}

func (c *Client) ListMultipartUploadsWithContext(ctx context.Context, bucket, prefix, keyMarker, uploadIDMarker, delimiter string, maxUploads int) (result models.ListMultipartUploadsResult, err error) {
	ml, e := c.mc.ListMultipartUploads(ctx, bucket, prefix, keyMarker, uploadIDMarker, delimiter, maxUploads)
	if e != nil {
		return result, e
	}
	// Convert minio to models
	output := models.ListMultipartUploadsResult{
		Bucket:             ml.Bucket,
		KeyMarker:          ml.KeyMarker,
		UploadIDMarker:     ml.UploadIDMarker,
		NextKeyMarker:      ml.NextKeyMarker,
		NextUploadIDMarker: ml.NextUploadIDMarker,
		EncodingType:       ml.EncodingType,
		MaxUploads:         ml.MaxUploads,
		IsTruncated:        ml.IsTruncated,
		Uploads:            []models.MultipartObjectInfo{},
		Prefix:             ml.Prefix,
		Delimiter:          ml.Delimiter,
		CommonPrefixes:     []models.CommonPrefix{},
	}
	for _, u := range ml.Uploads {
		output.Uploads = append(output.Uploads, models.MultipartObjectInfo{
			Initiated:    u.Initiated,
			Initiator:    u.Initiator,
			Owner:        u.Owner,
			StorageClass: u.StorageClass,
			Key:          u.Key,
			Size:         u.Size,
			UploadID:     u.UploadID,
			Err:          u.Err,
		})
	}
	for _, c := range ml.CommonPrefixes {
		output.CommonPrefixes = append(output.CommonPrefixes, models.CommonPrefix{Prefix: c.Prefix})
	}
	return output, nil
}

func (c *Client) ListObjectPartsWithContext(ctx context.Context, bucketName, objectName, uploadID string, partNumberMarker, maxParts int) (models.ListObjectPartsResult, error) {
	opp, er := c.mc.ListObjectParts(ctx, bucketName, objectName, uploadID, partNumberMarker, maxParts)
	if er != nil {
		return models.ListObjectPartsResult{}, er
	}
	lpi := models.ListObjectPartsResult{
		Bucket:               opp.Bucket,
		Key:                  opp.Key,
		UploadID:             opp.UploadID,
		Initiator:            opp.Initiator,
		Owner:                opp.Owner,
		StorageClass:         opp.StorageClass,
		PartNumberMarker:     opp.PartNumberMarker,
		NextPartNumberMarker: opp.NextPartNumberMarker,
		MaxParts:             opp.MaxParts,
		IsTruncated:          opp.IsTruncated,
		EncodingType:         opp.EncodingType,
	}
	for _, part := range lpi.ObjectParts {
		lpi.ObjectParts = append(lpi.ObjectParts, models.MultipartObjectPart{
			PartNumber:   part.PartNumber,
			LastModified: part.LastModified,
			ETag:         part.ETag,
			Size:         part.Size,
		})
	}
	return lpi, nil
}

func (c *Client) CompleteMultipartUploadWithContext(ctx context.Context, bucket, object, uploadID string, parts []models.MultipartObjectPart) (string, error) {
	cparts := make([]minio.CompletePart, len(parts))
	for i, p := range parts {
		cparts[i] = minio.CompletePart{
			PartNumber: p.PartNumber,
			ETag:       p.ETag,
		}
	}
	return c.mc.CompleteMultipartUpload(ctx, bucket, object, uploadID, cparts, minio.PutObjectOptions{})
}

func (c *Client) PutObjectPartWithContext(ctx context.Context, bucket, object, uploadID string, partID int, data io.Reader, size int64, md5Base64, sha256Hex string) (models.MultipartObjectPart, error) {
	pp, e := c.mc.PutObjectPart(ctx, bucket, object, uploadID, partID, data, size, md5Base64, sha256Hex, nil)
	if e != nil {
		return models.MultipartObjectPart{}, e
	}
	return models.MultipartObjectPart{
		PartNumber:   pp.PartNumber,
		LastModified: pp.LastModified,
		ETag:         pp.ETag,
		Size:         pp.Size,
	}, nil
}

func (c *Client) AbortMultipartUploadWithContext(ctx context.Context, bucket, object, uploadID string) error {
	return c.mc.AbortMultipartUpload(ctx, bucket, object, uploadID)
}

func (c *Client) CopyObject(sourceBucket, sourceObject, destBucket, destObject string, metadata map[string]string) (models.ObjectInfo, error) {
	oi, e := c.mc.CopyObject(context.Background(), sourceBucket, sourceObject, destBucket, destObject, metadata, minio.CopySrcOptions{}, minio.PutObjectOptions{})
	if e != nil {
		return models.ObjectInfo{}, e
	}
	return minioInfoToModelsInfo(oi), e
}

func (c *Client) CopyObjectWithProgress(sourceBucket, sourceObject, destBucket, destObject string, srcMeta map[string]string, metadata map[string]string, progress io.Reader) error {
	srcOptions := minio.CopySrcOptions{
		Bucket: sourceBucket,
		Object: sourceObject,
	}
	destOptions := minio.PutObjectOptions{
		Progress:     progress,
		UserMetadata: metadata,
	}
	/*
		destinationInfo, _ := NewDe(destBucket, destObject, nil, metadata)
		sourceInfo := minio.NewSourceInfo(sourceBucket, sourceObject, nil)
		// Add request Headers to SrcInfo (authentication, etc)
		for k, v := range srcMeta {
			sourceInfo.Headers.Set(k, v)
		}
	*/

	_, e := c.mc.CopyObject(context.Background(), sourceBucket, sourceObject, destBucket, destObject, srcMeta, srcOptions, destOptions)
	return e
}

func (c *Client) CopyObjectPartWithContext(ctx context.Context, srcBucket, srcObject, destBucket, destObject string, uploadID string, partID int, startOffset, length int64, metadata map[string]string) (p models.MultipartObjectPart, err error) {
	oi, e := c.mc.CopyObjectPart(ctx, srcBucket, srcObject, destBucket, destObject, uploadID, partID, startOffset, length, metadata)
	if e != nil {
		return models.MultipartObjectPart{}, e
	}
	return models.MultipartObjectPart{
		PartNumber: oi.PartNumber,
		ETag:       oi.ETag,
	}, e
}

func (c *Client) CopyObjectPart(srcBucket, srcObject, destBucket, destObject string, uploadID string, partID int, startOffset, length int64, metadata map[string]string) (p models.MultipartObjectPart, err error) {
	oi, e := c.mc.CopyObjectPart(context.Background(), srcBucket, srcObject, destBucket, destObject, uploadID, partID, startOffset, length, metadata)
	if e != nil {
		return models.MultipartObjectPart{}, e
	}
	return models.MultipartObjectPart{
		PartNumber: oi.PartNumber,
		ETag:       oi.ETag,
	}, e
}

// ListenBucketNotification hooks to events - Not part of the interface
func (c *Client) ListenBucketNotification(ctx context.Context, bucketName, prefix, suffix string, events []string) <-chan notification.Info {
	return c.mc.ListenBucketNotification(ctx, bucketName, prefix, suffix, events)
}

func readMetaToMinioOpts(meta models.ReadMeta) minio.GetObjectOptions {
	opt := minio.GetObjectOptions{}
	for k, v := range meta {
		opt.Set(k, v)
	}
	return opt
}

func putMetaToMinioOpts(meta models.PutMeta) minio.PutObjectOptions {
	opt := minio.PutObjectOptions{
		UserMetadata:            meta.UserMetadata,
		Progress:                meta.Progress,
		ContentType:             meta.ContentType,
		ContentEncoding:         meta.ContentEncoding,
		ContentDisposition:      meta.ContentDisposition,
		ContentLanguage:         meta.ContentLanguage,
		CacheControl:            meta.CacheControl,
		ServerSideEncryption:    nil,
		NumThreads:              meta.NumThreads,
		StorageClass:            meta.StorageClass,
		WebsiteRedirectLocation: meta.WebsiteRedirectLocation,
	}
	return opt
}

func minioInfoToModelsInfo(oi minio.ObjectInfo) models.ObjectInfo {
	return models.ObjectInfo{
		ETag:         oi.ETag,
		Key:          oi.Key,
		LastModified: oi.LastModified,
		Size:         oi.Size,
		ContentType:  oi.ContentType,
		Metadata:     oi.Metadata,
		Owner:        &models.ObjectInfoOwner{DisplayName: oi.Owner.DisplayName, ID: oi.Owner.ID},
		StorageClass: oi.StorageClass,
		Err:          oi.Err,
	}
}

func minioUploadInfoToModelsInfo(oi minio.UploadInfo) models.ObjectInfo {
	return models.ObjectInfo{
		ETag:         oi.ETag,
		Key:          oi.Key,
		LastModified: oi.LastModified,
		Size:         oi.Size,
	}
}
