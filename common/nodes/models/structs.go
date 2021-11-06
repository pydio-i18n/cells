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

package models

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pydio/cells/common"
)

type PutRequestData struct {
	Size              int64
	Md5Sum            []byte
	Sha256Sum         []byte
	Metadata          map[string]string
	MultipartUploadID string
	MultipartPartID   int
}

// MetaContentType looks for Content-Type or content-type key in metadata
func (p *PutRequestData) MetaContentType() string {
	if p.Metadata == nil {
		return ""
	}
	if c, o := p.Metadata[common.XContentType]; o {
		return c
	}
	if c, o := p.Metadata[strings.ToLower(common.XContentType)]; o {
		return c
	}
	return ""
}

// ContentTypeUnknown checks if cType is empty or generic "application/octet-stream"
func (p *PutRequestData) ContentTypeUnknown() bool {
	cType := p.MetaContentType()
	return cType == "" || cType == "application/octet-stream"
}

type GetRequestData struct {
	StartOffset int64
	Length      int64
	VersionId   string
}

type CopyRequestData struct {
	Metadata     map[string]string
	SrcVersionId string
	Progress     io.Reader
}

type MultipartRequestData struct {
	Metadata map[string]string

	ListKeyMarker      string
	ListUploadIDMarker string
	ListDelimiter      string
	ListMaxUploads     int
}

// S3ObjectInfo container for object metadata.
type S3ObjectInfo struct {
	// An ETag is optionally set to md5sum of an object.  In case of multipart objects,
	// ETag is of the form MD5SUM-N where MD5SUM is md5sum of all individual md5sums of
	// each parts concatenated into one string.
	ETag string `json:"etag"`

	Key          string    `json:"name"`         // Name of the object
	LastModified time.Time `json:"lastModified"` // Date and time the object was last modified.
	Size         int64     `json:"size"`         // Size in bytes of the object.
	ContentType  string    `json:"contentType"`  // A standard MIME type describing the format of the object data.

	// Collection of additional metadata on the object.
	// eg: x-amz-meta-*, content-encoding etc.
	Metadata http.Header `json:"metadata" xml:"-"`

	// Owner name.
	Owner struct {
		DisplayName string `json:"name"`
		ID          string `json:"id"`
	} `json:"owner"`

	// The class of storage used to store the object.
	StorageClass string `json:"storageClass"`

	// Error
	Err error `json:"-"`
}

// MultipartObjectPart container for particular part of an object.
// Can be used as CompletePart as well
type MultipartObjectPart struct {
	// Part number identifies the part.
	PartNumber int

	// Date and time the part was uploaded.
	LastModified time.Time

	// Entity tag returned when the part was uploaded, usually md5sum
	// of the part.
	ETag string

	// Size of the uploaded part data.
	Size int64
}

// MultipartObjectInfo container for multipart object metadata.
type MultipartObjectInfo struct {
	// Date and time at which the multipart upload was initiated.
	Initiated time.Time `type:"timestamp" timestampFormat:"iso8601"`

	Initiator struct {
		ID          string
		DisplayName string
	}
	Owner struct {
		DisplayName string
		ID          string
	}
	// The type of storage to use for the object. Defaults to 'STANDARD'.
	StorageClass string

	// Key of the object for which the multipart upload was initiated.
	Key string

	// Size in bytes of the object.
	Size int64

	// Upload ID that identifies the multipart upload.
	UploadID string `xml:"UploadId"`

	// Error
	Err error
}

// ListMultipartUploadsResult container for ListMultipartUploads response
type ListMultipartUploadsResult struct {
	Bucket             string
	KeyMarker          string
	UploadIDMarker     string `xml:"UploadIdMarker"`
	NextKeyMarker      string
	NextUploadIDMarker string `xml:"NextUploadIdMarker"`
	EncodingType       string
	MaxUploads         int64
	IsTruncated        bool
	Uploads            []MultipartObjectInfo `xml:"Upload"`
	Prefix             string
	Delimiter          string
	// A response can contain CommonPrefixes only if you specify a delimiter.
	CommonPrefixes []CommonPrefix
}

// ListObjectPartsResult container for ListObjectParts response.
type ListObjectPartsResult struct {
	Bucket   string
	Key      string
	UploadID string `xml:"UploadId"`

	Initiator struct {
		ID          string
		DisplayName string
	}
	Owner struct {
		DisplayName string
		ID          string
	}

	StorageClass         string
	PartNumberMarker     int
	NextPartNumberMarker int
	MaxParts             int

	// Indicates whether the returned list of parts is truncated.
	IsTruncated bool
	ObjectParts []MultipartObjectPart `xml:"Part"`

	EncodingType string
}

type CommonPrefix struct {
	Prefix string
}
