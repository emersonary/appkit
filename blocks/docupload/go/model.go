package docupload

// Session is the Redis metadata for one account's upload of one file id.
type Session struct {
	AccountID      string
	FileID         string
	FileName       string
	FileUnixMs     int64
	OriginalSize   int64
	CompressedMD5  string
	CompressedSize int64
	BlockSize      int
	BlockCount     int
	MaxThreads     int
	ExpiresUnixMs  int64
	ContentType    string
	Source         string
}

// StartRequest opens or resumes an upload session.
type StartRequest struct {
	FileID         string
	FileName       string
	FileUnixMs     int64
	OriginalSize   int64
	CompressedMD5  string
	CompressedSize int64
	ExpiresUnixMs  int64
	ContentType    string
	Source         string
}

// StartResult tells the client how to slice and send the compressed body.
type StartResult struct {
	BlockSize     int32
	Pending       []int32
	MaxThreads    int32
	ExpiresUnixMs int64
}

// BlockRequest is one slice of the compressed body.
type BlockRequest struct {
	FileID   string
	Index    int32
	Payload  []byte
	BlockMD5 string
}

// BlockResult reports whether the slice was stored and which indexes remain.
type BlockResult struct {
	Success bool
	Pending []int32
}

// CompleteDoc is the checked original file plus the metadata captured at start.
type CompleteDoc struct {
	FileID         string
	FileName       string
	FileUnixMs     int64
	OriginalSize   int64
	CompressedMD5  string
	CompressedSize int64
	ExpiresUnixMs  int64
	ContentType    string
	Source         string
	AccountID      string
	Bytes          []byte
}
