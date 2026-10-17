// Package docupload stores resumable gzip uploads in Redis.
// The file id is the MD5 of the original bytes. Blocks are slices of the
// compressed body. GetCompleteDoc rebuilds and checks both digests.
package docupload
