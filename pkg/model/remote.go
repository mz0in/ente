package model

import (
	"cli-go/pkg/model/export"
	"fmt"
	"time"
)

type FileType int8

const (
	Image FileType = iota
	Video
	LivePhoto
	Unknown = 127
)

type RemoteFile struct {
	ID              int64                  `json:"id"`
	OwnerID         int64                  `json:"ownerID"`
	Key             EncString              `json:"key"`
	LastUpdateTime  int64                  `json:"lastUpdateTime"`
	FileNonce       string                 `json:"fileNonce"`
	ThumbnailNonce  string                 `json:"thumbnailNonce"`
	Metadata        map[string]interface{} `json:"metadata"`
	PrivateMetadata map[string]interface{} `json:"privateMetadata"`
	PublicMetadata  map[string]interface{} `json:"publicMetadata"`
	Info            Info                   `json:"info"`
}

type Info struct {
	FileSize      int64 `json:"fileSize,omitempty"`
	ThumbnailSize int64 `json:"thumbSize,omitempty"`
}

type RemoteAlbum struct {
	ID            int64                  `json:"id"`
	OwnerID       int64                  `json:"ownerID"`
	IsShared      bool                   `json:"isShared"`
	IsDeleted     bool                   `json:"isDeleted"`
	AlbumName     string                 `json:"albumName"`
	AlbumKey      EncString              `json:"albumKey"`
	PublicMeta    map[string]interface{} `json:"publicMeta"`
	PrivateMeta   map[string]interface{} `json:"privateMeta"`
	SharedMeta    map[string]interface{} `json:"sharedMeta"`
	LastUpdatedAt int64                  `json:"lastUpdatedAt"`
}

type AlbumFileEntry struct {
	FileID        int64 `json:"fileID"`
	AlbumID       int64 `json:"albumID"`
	IsDeleted     bool  `json:"isDeleted"`
	SyncedLocally bool  `json:"localSync"`
}

func (r *RemoteFile) GetFileType() FileType {
	value, ok := r.Metadata["fileType"]
	if !ok {
		panic("fileType not found in metadata")
	}
	switch value.(int8) {
	case 0:
		return Image
	case 1:
		return Video
	case 2:
		return LivePhoto
	}
	panic(fmt.Sprintf("invalid fileType %d", value.(int8)))
}

func (r *RemoteFile) GetFileHash() *string {
	value, ok := r.Metadata["hash"]
	if !ok {
		return nil
	}
	if str, ok := value.(string); ok {
		return &str
	}
	return nil
}

func (r *RemoteFile) GetTitle() string {
	if r.PublicMetadata != nil {
		if value, ok := r.PublicMetadata["editedName"]; ok {
			return value.(string)
		}
	}
	value, ok := r.Metadata["title"]
	if !ok {
		panic("title not found in metadata")
	}
	return value.(string)
}

func (r *RemoteFile) GetCaption() *string {
	if r.PublicMetadata != nil {
		if value, ok := r.PublicMetadata["caption"]; ok {
			if str, ok := value.(string); ok {
				return &str
			}
		}
	}
	return nil
}

func (r *RemoteFile) GetCreationTime() time.Time {

	if r.PublicMetadata != nil {
		if value, ok := r.PublicMetadata["editedTime"]; ok && value.(float64) != 0 {
			return time.UnixMicro(int64(value.(float64)))
		}
	}
	value, ok := r.Metadata["creationTime"]
	if !ok {
		panic("creationTime not found in metadata")
	}
	return time.UnixMicro(int64(value.(float64)))
}

func (r *RemoteFile) GetModificationTime() time.Time {
	value, ok := r.Metadata["modificationTime"]
	if !ok {
		panic("creationTime not found in metadata")
	}
	return time.UnixMicro(int64(value.(float64)))
}

func (r *RemoteFile) GetLatlong() *export.Location {
	if r.ID == 10698020 {
		fmt.Println("found 10698020")
	}
	if r.PublicMetadata != nil {
		// check if lat and long key exists
		if lat, ok := r.PublicMetadata["lat"]; ok {
			if long, ok := r.PublicMetadata["long"]; ok {
				if lat.(float64) == 0 && long.(float64) == 0 {
					return nil
				}
				return &export.Location{
					Latitude:  lat.(float64),
					Longitude: long.(float64),
				}
			}
		}
	}
	if lat, ok := r.Metadata["latitude"]; ok && lat != nil {
		if long, ok2 := r.Metadata["longitude"]; ok2 && long != nil {
			return &export.Location{
				Latitude:  lat.(float64),
				Longitude: long.(float64),
			}
		}
	}
	return nil
}
