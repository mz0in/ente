package pkg

import (
	"cli-go/pkg/mapper"
	"cli-go/pkg/model"
	"cli-go/utils/encoding"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"
)

func (c *ClICtrl) fetchRemoteCollections(ctx context.Context) error {
	lastSyncTime, err2 := c.GetInt64ConfigValue(ctx, model.CollectionsSyncKey)
	if err2 != nil {
		return err2
	}
	collections, err := c.Client.GetCollections(ctx, lastSyncTime)
	if err != nil {
		return fmt.Errorf("failed to get collections: %s", err)
	}
	maxUpdated := lastSyncTime
	for _, collection := range collections {
		if lastSyncTime == 0 && collection.IsDeleted {
			continue
		}
		album, mapErr := mapper.MapCollectionToAlbum(ctx, collection, c.KeyHolder)
		if mapErr != nil {
			return mapErr
		}
		if album.LastUpdatedAt > maxUpdated {
			maxUpdated = album.LastUpdatedAt
		}
		albumJson := encoding.MustMarshalJSON(album)
		putErr := c.PutValue(ctx, model.RemoteAlbums, []byte(strconv.FormatInt(album.ID, 10)), albumJson)
		if putErr != nil {
			return putErr
		}
	}
	if maxUpdated > lastSyncTime {
		err = c.PutConfigValue(ctx, model.CollectionsSyncKey, []byte(strconv.FormatInt(maxUpdated, 10)))
		if err != nil {
			return fmt.Errorf("failed to update last sync time: %s", err)
		}
	}
	return nil
}

func (c *ClICtrl) fetchRemoteFiles(ctx context.Context) error {
	albums, err := c.getRemoteAlbums(ctx)
	if err != nil {
		return err
	}
	for _, album := range albums {
		if album.IsDeleted {
			log.Printf("Skipping album %s as it is deleted", album.AlbumName)
			continue
		}

		lastSyncTime, lastSyncTimeErr := c.GetInt64ConfigValue(ctx, fmt.Sprintf(model.CollectionsFileSyncKeyFmt, album.ID))
		if lastSyncTimeErr != nil {
			return lastSyncTimeErr
		}

		isFirstSync := lastSyncTime == 0

		for {
			if lastSyncTime == album.LastUpdatedAt {
				break
			}
			if isFirstSync {
				log.Printf("First sync for album %s\n", album.AlbumName)
			} else {
				log.Printf("Syncing album %s\n from %s", album.AlbumName, time.UnixMicro(lastSyncTime))
			}
			if !isFirstSync {
				t := time.UnixMicro(lastSyncTime)
				log.Printf("Fetching files for album %s from %v\n", album.AlbumName, t)
			}
			files, hasMore, err := c.Client.GetFiles(ctx, album.ID, lastSyncTime)
			if err != nil {
				return err
			}
			maxUpdated := lastSyncTime
			for _, file := range files {
				if file.UpdationTime > maxUpdated {
					maxUpdated = file.UpdationTime
				}
				if isFirstSync && file.IsDeleted {
					// on first sync, no need to sync delete markers
					continue
				}
				albumEntry := model.AlbumFileEntry{AlbumID: album.ID, FileID: file.ID, IsDeleted: file.IsDeleted, SyncedLocally: false}
				albumEntryJson := encoding.MustMarshalJSON(albumEntry)
				putErr := c.PutValue(ctx, model.RemoteAlbumEntries, []byte(fmt.Sprintf("%d:%d", album.ID, file.ID)), albumEntryJson)
				if putErr != nil {
					return putErr
				}
				if file.IsDeleted {
					continue
				}
				photoFile, err := mapper.MapApiFileToPhotoFile(ctx, album, file, c.KeyHolder)
				if err != nil {
					return err
				}
				fileJson := encoding.MustMarshalJSON(photoFile)
				// todo: use batch put
				putErr = c.PutValue(ctx, model.RemoteFiles, []byte(strconv.FormatInt(file.ID, 10)), fileJson)
				if putErr != nil {
					return putErr
				}
			}
			if !hasMore {
				maxUpdated = album.LastUpdatedAt
			}
			if (maxUpdated > lastSyncTime) || !hasMore {
				log.Printf("Updating last sync time for album %s to %s\n", album.AlbumName, time.UnixMicro(maxUpdated))
				err = c.PutConfigValue(ctx, fmt.Sprintf(model.CollectionsFileSyncKeyFmt, album.ID), []byte(strconv.FormatInt(maxUpdated, 10)))
				if err != nil {
					return fmt.Errorf("failed to update last sync time: %s", err)
				} else {
					lastSyncTime = maxUpdated
				}
			}
		}
	}
	return nil
}

func (c *ClICtrl) getRemoteAlbums(ctx context.Context) ([]model.RemoteAlbum, error) {
	albums := make([]model.RemoteAlbum, 0)
	albumBytes, err := c.GetAllValues(ctx, model.RemoteAlbums)
	if err != nil {
		return nil, err
	}
	for _, albumJson := range albumBytes {
		album := model.RemoteAlbum{}
		err = json.Unmarshal(albumJson, &album)
		if err != nil {
			return nil, err
		}
		albums = append(albums, album)
	}
	return albums, nil
}

func (c *ClICtrl) getRemoteFiles(ctx context.Context) ([]model.RemoteFile, error) {
	files := make([]model.RemoteFile, 0)
	fileBytes, err := c.GetAllValues(ctx, model.RemoteFiles)
	if err != nil {
		return nil, err
	}
	for _, fileJson := range fileBytes {
		file := model.RemoteFile{}
		err = json.Unmarshal(fileJson, &file)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, nil
}

func (c *ClICtrl) getRemoteAlbumEntries(ctx context.Context) ([]*model.AlbumFileEntry, error) {
	entries := make([]*model.AlbumFileEntry, 0)
	entryBytes, err := c.GetAllValues(ctx, model.RemoteAlbumEntries)
	if err != nil {
		return nil, err
	}
	for _, entryJson := range entryBytes {
		entry := &model.AlbumFileEntry{}
		err = json.Unmarshal(entryJson, &entry)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
