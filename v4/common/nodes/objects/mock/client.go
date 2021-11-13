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

package mock

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/pydio/cells/v4/common/nodes/models"
)

func New(buckets ...string) *Client {
	c := &Client{
		Buckets: map[string]map[string][]byte{},
	}
	for _, b := range buckets {
		c.Buckets[b] = make(map[string][]byte)
	}
	return c
}

type Client struct {
	Buckets map[string]map[string][]byte
}

func (c *Client) ListBucketsWithContext(ctx context.Context) (bb []models.BucketInfo, e error) {
	for b := range c.Buckets {
		bb = append(bb, models.BucketInfo{Name: b, CreationDate: time.Now()})
	}
	return
}

func (c *Client) MakeBucketWithContext(ctx context.Context, bucketName string, location string) (err error) {
	if _, ok := c.Buckets[bucketName]; ok {
		return fmt.Errorf("bucket already exists")
	}
	c.Buckets[bucketName] = map[string][]byte{}
	return nil
}

func (c *Client) RemoveBucketWithContext(ctx context.Context, bucketName string) error {
	if _, ok := c.Buckets[bucketName]; !ok {
		return fmt.Errorf("bucket not found")
	}
	delete(c.Buckets, bucketName)
	return nil
}

func (c *Client) GetObject(bucketName, objectName string, opts models.ReadMeta) (io.ReadCloser, models.ObjectInfo, error) {
	bucket, ok := c.Buckets[bucketName]
	if !ok {
		return nil, models.ObjectInfo{}, fmt.Errorf("bucket not found")
	}
	if object, ok := bucket[objectName]; ok {
		return newReadCloser(object), models.ObjectInfo{Size: int64(len(object))}, nil
	} else {
		return nil, models.ObjectInfo{}, fmt.Errorf("object not found")
	}
}

func (c *Client) GetObjectWithContext(ctx context.Context, bucketName, objectName string, opts models.ReadMeta) (io.ReadCloser, error) {
	rc, _, e := c.GetObject(bucketName, objectName, opts)
	return rc, e
}

func (c *Client) StatObject(bucketName, objectName string, opts models.ReadMeta) (models.ObjectInfo, error) {
	bucket, ok := c.Buckets[bucketName]
	if !ok {
		return models.ObjectInfo{}, fmt.Errorf("bucket not found")
	}
	if object, ok := bucket[objectName]; ok {
		return models.ObjectInfo{Size: int64(len(object))}, nil
	} else {
		return models.ObjectInfo{}, fmt.Errorf("object not found")
	}
}

func (c *Client) PutObject(bucketName, objectName string, data io.Reader, size int64, md5Base64, sha256Hex string, metadata models.ReadMeta) (models.ObjectInfo, error) {
	bucket, ok := c.Buckets[bucketName]
	if !ok {
		return models.ObjectInfo{}, fmt.Errorf("bucket not found")
	}
	bucket[objectName], _ = io.ReadAll(data)
	return models.ObjectInfo{Size: int64(len(bucket[objectName]))}, fmt.Errorf("not.implemented")
}

func (c *Client) PutObjectWithContext(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts models.PutMeta) (n int64, err error) {
	oi, e := c.PutObject(bucketName, objectName, reader, objectSize, "", "", opts.UserMetadata)
	return oi.Size, e
}

func (c *Client) RemoveObjectWithContext(ctx context.Context, bucketName, objectName string) error {
	bucket, ok := c.Buckets[bucketName]
	if !ok {
		return fmt.Errorf("bucket not found")
	}
	if _, ok := bucket[objectName]; !ok {
		return fmt.Errorf("object not found")
	}
	delete(bucket, objectName)
	return nil
}

func (c *Client) ListObjectsWithContext(ctx context.Context, bucketName, prefix, marker, delimiter string, maxKeys int) (result models.ListBucketResult, err error) {
	bucket, ok := c.Buckets[bucketName]
	if !ok {
		return result, fmt.Errorf("bucket not found")
	}
	for objName, data := range bucket {
		result.Contents = append(result.Contents, models.ObjectInfo{
			Key:          objName,
			LastModified: time.Now(),
			Size:         int64(len(data)),
		})
	}
	return result, nil
}

func (c *Client) NewMultipartUploadWithContext(ctx context.Context, bucket, object string, opts models.PutMeta) (uploadID string, err error) {
	return "", fmt.Errorf("not.implemented")
}

func (c *Client) ListMultipartUploadsWithContext(ctx context.Context, bucket, prefix, keyMarker, uploadIDMarker, delimiter string, maxUploads int) (result models.ListMultipartUploadsResult, err error) {
	return result, fmt.Errorf("not.implemented")
}

func (c *Client) ListObjectPartsWithContext(ctx context.Context, bucketName, objectName, uploadID string, partNumberMarker, maxParts int) (models.ListObjectPartsResult, error) {
	return models.ListObjectPartsResult{}, fmt.Errorf("not.implemented")
}

func (c *Client) CompleteMultipartUploadWithContext(ctx context.Context, bucket, object, uploadID string, parts []models.MultipartObjectPart) (string, error) {
	return "", fmt.Errorf("not.implemented")
}

func (c *Client) PutObjectPartWithContext(ctx context.Context, bucket, object, uploadID string, partID int, data io.Reader, size int64, md5Base64, sha256Hex string) (models.MultipartObjectPart, error) {
	return models.MultipartObjectPart{}, fmt.Errorf("not.implemented")
}

func (c *Client) AbortMultipartUploadWithContext(ctx context.Context, bucket, object, uploadID string) error {
	return fmt.Errorf("not.implemented")
}

func (c *Client) CopyObject(sourceBucket, sourceObject, destBucket, destObject string, metadata map[string]string) (models.ObjectInfo, error) {
	return models.ObjectInfo{}, fmt.Errorf("not.implemented")
}

func (c *Client) CopyObjectWithProgress(sourceBucket, sourceObject, destBucket, destObject string, srcMeta map[string]string, metadata map[string]string, progress io.Reader) error {
	return fmt.Errorf("not.implemented")
}

func (c *Client) CopyObjectPartWithContext(ctx context.Context, srcBucket, srcObject, destBucket, destObject string, uploadID string, partID int, startOffset, length int64, metadata map[string]string) (p models.MultipartObjectPart, err error) {
	return p, fmt.Errorf("not.implemented")
}

func (c *Client) CopyObjectPart(srcBucket, srcObject, destBucket, destObject string, uploadID string, partID int, startOffset, length int64, metadata map[string]string) (p models.MultipartObjectPart, err error) {
	return p, fmt.Errorf("not.implemented")
}

type mockReadCloser struct {
	*bytes.Buffer
}

func newReadCloser(bb []byte) *mockReadCloser {
	return &mockReadCloser{Buffer: bytes.NewBuffer(bb)}
}

func (m *mockReadCloser) Close() error {
	return nil
}
