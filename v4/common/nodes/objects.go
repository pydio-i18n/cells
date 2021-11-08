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

package nodes

import (
	"context"
	"io"

	"github.com/pydio/cells/v4/common/nodes/objects/mock"

	"github.com/pydio/cells/v4/common/nodes/models"
	"github.com/pydio/cells/v4/x/configx"
)

type StorageClient interface {
	ListBucketsWithContext(ctx context.Context) ([]models.BucketInfo, error)
	MakeBucketWithContext(ctx context.Context, bucketName string, location string) (err error)
	RemoveBucketWithContext(ctx context.Context, bucketName string) error

	GetObject(bucketName, objectName string, opts models.ReadMeta) (io.ReadCloser, models.ObjectInfo, error)
	GetObjectWithContext(ctx context.Context, bucketName, objectName string, opts models.ReadMeta) (io.ReadCloser, error)

	StatObject(bucketName, objectName string, opts models.ReadMeta) (models.ObjectInfo, error)
	PutObject(bucket, object string, data io.Reader, size int64, md5Base64, sha256Hex string, metadata models.ReadMeta) (models.ObjectInfo, error)
	PutObjectWithContext(ctx context.Context, bucketName, objectName string, reader io.Reader, objectSize int64,
		opts models.PutMeta) (n int64, err error)
	RemoveObjectWithContext(ctx context.Context, bucketName, objectName string) error
	ListObjectsWithContext(ctx context.Context, bucket, prefix, marker, delimiter string, maxKeys int) (result models.ListBucketResult, err error)

	NewMultipartUploadWithContext(ctx context.Context, bucket, object string, opts models.PutMeta) (uploadID string, err error)
	ListMultipartUploadsWithContext(ctx context.Context, bucket, prefix, keyMarker, uploadIDMarker, delimiter string, maxUploads int) (result models.ListMultipartUploadsResult, err error)
	ListObjectPartsWithContext(ctx context.Context, bucketName, objectName, uploadID string, partNumberMarker, maxParts int) (models.ListObjectPartsResult, error)
	CompleteMultipartUploadWithContext(ctx context.Context, bucket, object, uploadID string, parts []models.MultipartObjectPart) (string, error)
	PutObjectPartWithContext(ctx context.Context, bucket, object, uploadID string, partID int, data io.Reader, size int64, md5Base64, sha256Hex string) (models.MultipartObjectPart, error)
	AbortMultipartUploadWithContext(ctx context.Context, bucket, object, uploadID string) error

	CopyObject(sourceBucket, sourceObject, destBucket, destObject string, metadata map[string]string) (models.ObjectInfo, error)
	CopyObjectWithProgress(sourceBucket, sourceObject, destBucket, destObject string, srcMeta map[string]string, metadata map[string]string, progress io.Reader) error
	CopyObjectPartWithContext(ctx context.Context, srcBucket, srcObject, destBucket, destObject string, uploadID string,
		partID int, startOffset, length int64, metadata map[string]string) (p models.MultipartObjectPart, err error)
	CopyObjectPart(srcBucket, srcObject, destBucket, destObject string, uploadID string,
		partID int, startOffset, length int64, metadata map[string]string) (p models.MultipartObjectPart, err error)

	//ListObjects(bucket, prefix, marker, delimiter string, maxKeys int) (result minio.ListBucketResult, err error)
	//ListObjectsV2(bucketName, objectPrefix, continuationToken string, fetchOwner bool, delimiter string, maxkeys int, startAfter string) (minio.ListBucketV2Result, error)
	//NewMultipartUpload(bucket, object string, opts minio.PutObjectOptions) (uploadID string, err error)
	//ListMultipartUploads(bucket, prefix, keyMarker, uploadIDMarker, delimiter string, maxUploads int) (result minio.ListMultipartUploadsResult, err error)
	//ListObjectParts(bucket, object, uploadID string, partNumberMarker int, maxParts int) (result minio.ListObjectPartsResult, err error)
	//CompleteMultipartUpload(bucket, object, uploadID string, parts []minio.CompletePart) (string, error)
	//PutObjectPart(bucket, object, uploadID string, partID int, data io.Reader, size int64, md5Base64, sha256Hex string, sse encrypt.ServerSide) (minio.ObjectPart, error)
	//AbortMultipartUpload(bucket, object, uploadID string) error

}

func NewStorageClient(cfg configx.Values) (StorageClient, error) {
	return &mock.Client{}, nil
	/*
		ep := cfg.Val("endpoint").String()
		key := cfg.Val("key").String()
		secret := cfg.Val("secret").String()
		secure := cfg.Val("secure").Bool()
		return mc.New(ep, key, secret, secure)
	*/
}
