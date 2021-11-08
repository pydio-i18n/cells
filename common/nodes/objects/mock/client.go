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
	"context"
	"fmt"
	"io"

	"github.com/pydio/cells/common/nodes/models"
)

type Client struct {
}

func (c *Client) ListBucketsWithContext(ctx context.Context) ([]models.BucketInfo, error) {
	panic("implement me")
}

func (c *Client) MakeBucketWithContext(ctx context.Context, bucketName string, location string) (err error) {
	panic("implement me")
}

func (c *Client) RemoveBucketWithContext(ctx context.Context, bucketName string) error {
	panic("implement me")
}

func (c *Client) GetObject(bucketName, objectName string, opts models.ReadMeta) (io.ReadCloser, models.ObjectInfo, error) {
	return nil, models.ObjectInfo{}, fmt.Errorf("not.implemented")
}

func (c *Client) GetObjectWithContext(ctx context.Context, bucketName, objectName string, opts models.ReadMeta) (io.ReadCloser, error) {
	return nil, fmt.Errorf("not.implemented")
}

func (c *Client) StatObject(bucketName, objectName string, opts models.ReadMeta) (models.ObjectInfo, error) {
	return models.ObjectInfo{}, fmt.Errorf("not.implemented")
}

func (c *Client) PutObject(bucket, object string, data io.Reader, size int64, md5Base64, sha256Hex string, metadata models.ReadMeta) (models.ObjectInfo, error) {
	return models.ObjectInfo{}, fmt.Errorf("not.implemented")
}

func (c *Client) PutObjectWithContext(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64, opts models.PutMeta) (n int64, err error) {
	return 0, fmt.Errorf("not.implemented")
}

func (c *Client) RemoveObjectWithContext(ctx context.Context, bucketName, objectName string) error {
	return fmt.Errorf("not.implemented")
}

func (c *Client) ListObjectsWithContext(ctx context.Context, bucket, prefix, marker, delimiter string, maxKeys int) (result models.ListBucketResult, err error) {
	return result, fmt.Errorf("not.implemented")
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
