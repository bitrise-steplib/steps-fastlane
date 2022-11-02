package cache

import (
	"strings"
)

type skipReason int

const (
	reasonKeyNotDynamic skipReason = iota
	reasonNoRestore
	reasonRestoreSameUniqueKey
	reasonRestoreSameKeyNotUnique
	reasonNoRestoreThisKey
	reasonNewArchiveChecksumMatch
	reasonNewArchiveChecksumMismatch
)

func (r skipReason) String() string {
	switch r {
	case reasonKeyNotDynamic:
		return "key_not_dynamic"
	case reasonNoRestore:
		return "no_restore_found"
	case reasonRestoreSameUniqueKey:
		return "restore_same_unique_key"
	case reasonRestoreSameKeyNotUnique:
		return "restore_same_key_not_unique"
	case reasonNoRestoreThisKey:
		return "no_restore_with_this_key"
	case reasonNewArchiveChecksumMatch:
		return "new_archive_checksum_match"
	case reasonNewArchiveChecksumMismatch:
		return "new_archive_checksum_mismatch"
	default:
		return "unknown"
	}
}

func (r skipReason) description() string {
	switch r {
	case reasonKeyNotDynamic:
		return "key is not dynamic; the expectation is that the same key is used for saving different cache contents over and over"
	case reasonNoRestore:
		return "no cache was restored in the workflow, creating a new cache entry"
	case reasonRestoreSameUniqueKey:
		return "a cache with the same key was restored in the workflow, new cache would have the same content"
	case reasonRestoreSameKeyNotUnique:
		return "a cache with the same key was restored in the workflow, but contents might have changed since then"
	case reasonNoRestoreThisKey:
		return "there was no cache restore in the workflow with this key, but was for other(s)"
	case reasonNewArchiveChecksumMatch:
		return "new cache archive is the same as the restored one"
	case reasonNewArchiveChecksumMismatch:
		return "new cache archive doesn't match the restored one"
	default:
		return "unrecognized skipReason"
	}
}

func (s *saver) canSkipSave(keyTemplate, evaluatedKey string, isKeyUnique bool) (bool, skipReason) {
	if keyTemplate == evaluatedKey {
		return false, reasonKeyNotDynamic
	}

	cacheHits := s.getCacheHits()
	if len(cacheHits) == 0 {
		return false, reasonNoRestore
	}

	if _, ok := cacheHits[evaluatedKey]; ok {
		if isKeyUnique {
			return true, reasonRestoreSameUniqueKey
		} else {
			return false, reasonRestoreSameKeyNotUnique
		}
	}

	return false, reasonNoRestoreThisKey
}

func (s *saver) canSkipUpload(newCacheKey, newCacheChecksum string) (bool, skipReason) {
	cacheHits := s.getCacheHits()

	if len(cacheHits) == 0 {
		return false, reasonNoRestore
	}

	checksumForNewKey, ok := cacheHits[newCacheKey]
	if !ok {
		return false, reasonNoRestoreThisKey
	}
	if checksumForNewKey == newCacheChecksum {
		return true, reasonNewArchiveChecksumMatch
	}

	return false, reasonNewArchiveChecksumMismatch
}

// Returns cache hit information exposed by previous restore cache steps.
// The returned map's key is the restored cache key, and the value is the checksum of the cache archive
func (s *saver) getCacheHits() map[string]string {
	cacheHits := map[string]string{}
	for _, e := range s.envRepo.List() {
		envParts := strings.SplitN(e, "=", 2)
		if len(envParts) < 2 {
			continue
		}
		envKey := envParts[0]
		envValue := envParts[1]

		if strings.HasPrefix(envKey, cacheHitEnvVarPrefix) {
			cacheKey := strings.TrimPrefix(envKey, cacheHitEnvVarPrefix)
			cacheHits[cacheKey] = envValue
		}
	}
	return cacheHits
}

func (s *saver) logOtherHits() {
	otherKeys := []string{}
	for k := range s.getCacheHits() {
		otherKeys = append(otherKeys, k)
	}
	if len(otherKeys) == 0 {
		return
	}
	s.logger.Printf("Other restored cache keys:")
	s.logger.Printf(strings.Join(otherKeys, "\n"))
}
