// Code generated by vfsgen; DO NOT EDIT.

package fixtures

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	pathpkg "path"
	"time"
)

// Fixtures statically implements the virtual filesystem provided to vfsgen.
var Fixtures = func() http.FileSystem {
	fs := vfsgen۰FS{
		"/": &vfsgen۰DirInfo{
			name:    "/",
			modTime: time.Date(2022, 6, 14, 6, 24, 16, 242233136, time.UTC),
		},
		"/agent_templates.yaml": &vfsgen۰CompressedFileInfo{
			name:             "agent_templates.yaml",
			modTime:          time.Date(2022, 6, 13, 4, 4, 45, 715906636, time.UTC),
			uncompressedSize: 4367,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\xe4\x57\xc1\x8e\x9b\x30\x10\xbd\xf7\x2b\x46\xdc\x43\xe9\xa9\x12\xb7\x88\x54\x6a\x0f\xdb\xae\xb2\xdb\x0f\x70\xcd\x40\x2d\x19\x1b\xd9\x43\xb4\x68\xb5\xff\x5e\x19\x4c\x0a\x0d\x44\xa0\x94\x4d\x56\x9b\x1b\xd8\x33\xcc\xcc\x7b\xcf\x79\x16\x84\x85\x8d\x3f\x00\x6c\xa0\x40\x62\x29\x23\xe6\x9e\xdc\x4f\xb1\x02\x63\x28\x99\x61\xb2\xb2\x1b\x8b\x8a\x4c\xbd\x29\x11\x8d\x50\xf9\xc6\xa2\x39\xa0\xf1\x3b\x53\x61\x4b\xc9\xea\xef\x4d\x40\x70\xdf\x46\xc0\x43\x13\x01\xf7\x6d\x04\x3c\x34\x11\x41\x13\x62\x4b\xe4\xdd\x67\x84\xca\x0c\xdb\x63\x36\xf5\x29\xbf\x8d\x55\xa4\xf7\x98\x0b\x4b\x68\x62\x20\x53\x61\x6f\x61\x5b\x96\x46\x1f\x70\xf0\x5e\xe4\x4a\x1b\xbc\xab\x24\x89\x52\xe2\x68\x28\x61\x51\x4a\x46\xf8\x58\x97\xae\xf0\x7e\x85\x00\xbf\xb5\x25\x1b\x83\x7f\x72\x03\x72\x6f\x62\x08\x9e\x9f\x21\x6c\x4b\xf4\xad\x7d\xd5\x96\xe0\xe5\x25\x38\x6e\x05\xa0\x36\xa3\x5b\x71\xc9\xbf\x3c\x11\x1a\xc5\x64\x70\x9a\x6d\xd8\x74\xfc\x39\x8a\x3e\x9d\x49\xf4\x4d\xcd\x4c\x14\x1e\x1f\x6b\x4b\x58\x34\x79\x61\x85\xc4\xa1\x3d\xf0\x90\xcb\xca\xcd\x36\x94\x9a\x33\xb9\xb4\x05\xa1\x92\x36\xfc\xd1\xa3\x11\x43\xd0\xad\xe9\x8a\x7e\x64\xe3\xcb\x0b\xf9\xca\xa5\x40\x45\x37\x4c\xbe\xa4\x29\x70\xe5\xa1\x70\x6d\x70\x63\x50\xb2\x7a\x8e\x80\x13\x6d\x10\xf6\x6e\x37\x0c\xa5\x21\xd9\x2f\x94\x36\x3e\x62\xec\xd3\x87\x29\x1e\x3e\x72\xad\x14\x72\xd2\xfd\xd2\x5c\xd7\x73\x84\xff\xb7\xbc\xeb\x8a\x7e\x5c\xf3\xae\xba\x66\x1a\x6d\xc8\x72\xd5\xaf\x0f\x2a\xcb\x3b\x92\x9f\x07\x29\xc5\x8c\x55\x92\x9a\x6e\xfe\x1f\x3e\x19\x93\x76\x6d\x61\x9c\x3d\x95\xdf\x00\x42\x95\x9d\x23\xba\x9f\x16\xcd\x02\xcd\x79\x38\x5d\xd4\x00\xcd\x7f\xf7\xb9\x8f\x8f\xca\x72\x52\xc4\x5b\x47\xa8\x5e\xc8\x14\xe7\x96\xb1\xa7\x3f\x85\x57\x90\xf8\x9d\x78\xc2\x74\x19\x81\xdc\x28\x6f\x8b\x3e\x3c\x5d\xc4\xa0\x64\xe7\x4f\x6e\xd7\xc9\x02\x0e\x25\xbb\x13\x16\x35\xa8\xc2\x79\x58\xc7\xaa\xbb\x0e\xb2\xd3\xc0\x26\xbb\x0b\xa0\x1d\xf1\x43\x3c\x6d\xb9\xef\x0c\x4f\x74\x99\xb5\xea\x52\x8d\xb8\xb6\x08\x56\x49\x3d\xee\xdb\xa2\xeb\xfb\xb6\x21\x9d\xe6\xba\x94\x23\xdd\x4f\x6e\x19\x33\x89\x7b\xab\x86\xc3\x93\x36\xe9\x0e\xe4\x1b\x3d\x94\xfa\xc6\x63\x0e\x4c\xdb\xbc\xf7\x9f\x3e\xeb\x58\x9a\xf0\x2a\x97\xc0\xfb\x7a\x7e\xe5\x0d\xe2\xeb\xef\x41\x3c\x7d\x57\x57\xa8\x3f\x01\x00\x00\xff\xff\x2f\xb6\x4c\xa4\x0f\x11\x00\x00"),
		},
		"/relay_agent_template.yaml": &vfsgen۰CompressedFileInfo{
			name:             "relay_agent_template.yaml",
			modTime:          time.Date(2022, 6, 14, 6, 24, 16, 217233097, time.UTC),
			uncompressedSize: 3874,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\xd4\x57\x4b\x6f\xe3\x36\x10\xbe\xeb\x57\x10\xbe\xec\xa5\xca\x63\xd3\x5d\x2c\x08\xec\x41\x6b\x3b\x1b\xa3\x7e\x08\xb6\xd2\xee\xa2\x28\x8c\x09\x39\xb6\x89\x50\xa4\x40\x52\x4a\x84\x20\xff\xbd\xa0\x64\xd9\xb2\xe3\xb4\x69\xd6\x69\x51\x9e\xec\x79\x7e\xf3\x71\x6c\xce\x40\x26\x7e\x45\x63\x85\x56\x94\x40\x96\xd9\xd3\xe2\x3c\xb8\x15\x8a\x53\xd2\xc3\x4c\xea\x32\x45\xe5\x82\x14\x1d\x70\x70\x40\x03\x42\x24\xdc\xa0\xb4\xfe\x13\x21\x06\xb3\x10\x38\xf7\xbe\xc5\xfb\xd0\xa0\x84\x32\x84\xa5\xf7\x68\xb4\x4c\xe6\xd6\xa1\xa1\xa4\xf3\xf0\x40\x4e\x7a\xfa\x4e\x49\x0d\xbc\x07\x0e\x4e\xba\xb5\x6a\xd0\x23\x8f\x8f\x9d\x80\x10\x05\x29\x52\xd2\xb9\xcd\x6f\x30\xcc\x8c\xbe\x5f\x87\x0a\x5f\xe4\x68\x33\x60\x48\x49\x06\x06\x64\x6e\x43\x5b\x5a\x87\x69\x60\x33\x64\x1e\x69\x66\xf4\xd2\xa0\xb5\x3d\x04\x2e\x85\xc2\x19\x32\xad\xb8\xa5\xe4\xfc\xd3\xd9\x59\x50\x21\x95\x82\x81\x17\x54\xdf\x0a\xe1\x09\xb9\x12\xd6\x69\x53\x0e\x45\x2a\x1c\x25\xe7\xde\xd0\xa2\x44\xe6\xb4\xa9\xcb\x4f\xc1\xb1\xd5\xb0\xc5\x07\xf1\x14\xbe\xaa\x06\x87\x69\x26\xc1\xe1\x3a\x70\x8b\x6f\x7f\xe4\x4e\x8e\xd7\x67\x21\xa4\x61\xc4\x1f\xa6\x95\x03\xa1\xd0\x6c\x22\x87\x04\xcc\xb2\x95\x27\x24\x61\x98\x6a\x8e\x9f\x99\x14\xcd\xad\x36\x72\xa9\x97\xa1\xc4\x02\xe5\xe7\x8b\x8d\x1c\x55\xd1\x76\xae\x6f\x34\x9e\xf4\xe6\xe3\x68\xd4\xdf\x28\x08\x29\x40\xe6\x78\x69\x74\x4a\x5b\x42\x42\x16\x02\x25\x9f\xe2\x62\x57\xea\xcb\xdd\xf6\x68\x71\xbe\xa7\xac\x9c\x62\x70\x2b\xba\xa1\xed\xc4\x27\x7e\x16\xc7\x2c\x8e\xba\xff\x36\x98\xaa\x3d\x9f\x20\x1a\x45\xdf\xe6\xbd\x41\x34\x9c\xfd\x3d\x1a\xa6\xd5\x42\x2c\x47\x90\xfd\x82\xe5\x01\x50\xb7\x58\x52\x92\xc2\x7d\x4f\x80\xb4\x7b\xba\xe7\x7e\x57\x75\xc8\x97\x34\xcd\xf6\xe8\xcc\x09\xad\x40\x52\xe2\x4c\xfe\xb4\xa0\x69\x7f\x18\x7d\x9f\x77\x27\xe3\xcb\xc1\xd7\x51\x14\x1f\xbc\xf6\x1f\x87\xd2\x64\xf3\xd4\x4d\xae\x93\x79\x3c\x9d\x7c\xfb\x7e\x1c\x0a\x57\xce\x65\x36\xf6\xc8\x0e\x92\x58\x63\xae\x03\xfd\x33\x66\x76\xb0\xce\xa3\xeb\xe4\xaa\x3f\x4e\x06\xdd\x28\x19\x4c\xc6\xc7\x81\x5e\x61\x8b\x72\xb7\x3a\x32\xf2\xab\x24\x89\x8f\x4d\xf1\x5b\x30\xec\x71\xce\xfe\x17\xbd\x30\x9e\x1c\x13\xa6\xd2\x6f\x81\x31\x1a\x0e\x27\xbf\xcd\x07\xe3\x59\xbf\x7b\x3d\xed\xcf\xbf\x4c\x26\xc9\x2c\x99\x46\xf1\x71\x30\x83\x94\xfa\x6e\xa0\x2c\xb2\xdc\xe0\x17\xad\x9d\x75\x06\xb2\x23\x95\x20\x52\x58\xe2\xa1\x81\x63\xea\xe7\x93\x81\xd7\xee\xfc\x9d\x54\xf6\x71\x2e\x65\xac\xa5\x60\x25\x25\x83\xc5\x58\xbb\xd8\xa0\x6d\x3f\x79\xaf\x9e\x4e\xea\xe3\xd0\xa4\x42\x81\x07\x3c\x42\x6b\x7d\xc6\xea\x95\x38\xe5\x58\x9c\xb6\x94\xfe\x59\xfd\x2b\xa7\x35\xc4\x4b\x21\xb7\x05\x17\x5a\xe6\x29\x8e\x74\xae\xdc\xce\xdb\x9d\x7a\xc9\x3a\x0d\x3a\x76\xfa\x84\xc4\x63\xbc\x0c\x42\x09\xd7\x3d\x30\x48\x30\x9d\xa6\xa0\x78\x1b\x8f\x5d\xb5\x07\x08\xd6\xfa\x92\x4b\x3f\x63\x91\x50\x91\x8f\x1f\x3e\x5c\x7c\xdc\xbf\xca\x9b\xdc\x96\x37\xfa\x9e\x9e\x9f\x5c\x5c\xbc\xea\xda\x2c\xba\xb0\x4a\xb1\x7d\x1b\x0d\x5a\x9d\x1b\x86\x96\x92\x87\xc7\x8d\xb4\xea\x48\xe1\x4a\x5f\x11\xde\xbb\x76\xf3\x66\x46\x14\x42\xe2\x12\xf9\x5e\xbf\xbd\xdd\xd5\x66\x46\xe8\x0a\x8d\x04\x6b\xc7\xf5\xcf\x61\x3d\xe0\xae\xe7\xea\x90\x19\xe1\x04\x03\x19\x6c\xaa\x72\x60\x5c\x13\x2b\x92\x77\x50\x36\x35\x5b\xb6\x42\x9e\x4b\x34\x75\x24\x8e\x0b\xc8\xa5\x0b\x37\xe2\xe0\x30\x05\x5b\x7a\x5a\x98\xbf\x1a\x60\x18\xa3\x11\x9a\x6f\x67\xe9\xb3\xa0\xdd\x8f\x3b\xbd\xb0\xfe\x4b\x68\xf3\xb9\xce\x3f\xd2\x1c\x29\xf9\xf9\xfd\x59\x70\xec\x89\xe5\xc7\xa2\x04\x61\x18\x06\x7b\xb3\x5e\xbd\x17\x75\x9b\x6a\x82\x66\x44\x67\x8d\xe7\x4b\xd6\x9c\x66\x4a\xa3\xa4\xf3\xa9\x53\x6d\x1b\x12\x4a\x4b\xc9\xbb\xdf\x1f\x3a\x4e\xdf\xa2\xea\xd0\xa7\x41\x12\xaf\xf0\x01\x7e\xea\x00\xe7\xe6\x90\xc9\x0c\x95\x33\x65\xc4\xb9\xa9\xed\x50\xf1\x4c\x0b\xe5\x0e\xd9\x5e\x69\xeb\x7c\x17\xd4\x96\x9e\xa8\x83\x49\xd7\x3b\xc9\xd6\xb2\xd9\x52\x92\x67\x71\xb6\x0d\xbc\xcf\xe3\x1f\xef\xfe\xdb\xed\xf1\xa5\x3d\xf3\xfc\x12\xf9\x67\x00\x00\x00\xff\xff\xa9\x89\x3d\x4a\x22\x0f\x00\x00"),
		},
		"/relay_template.yaml": &vfsgen۰CompressedFileInfo{
			name:             "relay_template.yaml",
			modTime:          time.Date(2022, 6, 13, 4, 4, 45, 716906640, time.UTC),
			uncompressedSize: 6038,

			compressedContent: []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x02\xff\xec\x97\xc1\x6e\xdb\x38\x13\xc7\xef\x7a\x0a\x22\x77\x39\x09\x5a\x7c\x28\x04\xf4\xa0\xcf\x76\x12\x23\x89\x2c\xc8\x4a\x8a\x9e\x0c\x56\x9a\x38\x44\x28\x92\x4b\x8e\x9c\x35\x8a\xbe\xfb\x82\xb2\x24\x4b\xb6\xe4\x06\xae\x50\x6c\xb1\xc9\x29\x26\x67\x86\xbf\x19\x8f\xc9\xf9\xbb\xae\xeb\x50\xc5\x1e\x41\x1b\x26\x85\x47\xd6\x97\xce\x0b\x13\xa9\x47\x02\x9a\x81\x51\x34\x01\x27\x03\xa4\x29\x45\xea\x39\x84\x70\xfa\x0d\xb8\xb1\xff\x11\x92\x48\x81\x5a\x72\x57\x71\x2a\xc0\xab\x3e\x72\xd0\x6e\x46\x05\x5d\x81\x76\x08\x11\x34\x03\x8f\x28\xaa\x29\xcf\x8d\x6b\x36\x06\x21\x73\x1c\x7b\x68\xf7\xa9\xa1\x5d\x31\x08\x02\x1f\x25\xcf\x33\x18\x73\xca\xb2\x16\x41\x3b\xe2\x4b\xfe\x0d\x5c\xa5\xe5\xdf\x1b\x97\xe6\x29\xc3\xd2\xa0\x00\x3f\x38\xd7\x28\x48\x6c\x08\x9a\x24\x60\xcc\xbd\x4c\xa1\xc8\xc4\x25\x11\xd0\xf4\x8b\x66\x08\x73\x91\x80\x43\x88\x06\x23\x73\x9d\x40\x99\xa8\x86\xbf\x72\x30\x58\x7e\x22\xc4\xa0\xd4\x74\x05\x1e\xb9\xbc\x66\x0e\x21\xeb\x82\xd4\x86\xf3\xc8\x15\xe3\x50\x9e\x46\x9c\xfd\xda\x52\xa5\xcc\x79\x9d\xea\x04\x14\x97\x9b\x0c\x04\xf6\x56\x98\x2a\xd5\x95\xea\x91\x2a\xbc\x29\x7f\x0d\x8a\xb3\x84\x1a\x8f\x5c\x3a\x84\x28\x2d\x57\x1a\x8c\x99\x00\x4d\x39\x13\xb0\x80\x44\x8a\xd4\x6e\x7e\xba\xb8\x70\x08\x31\xc0\x21\x41\xa9\xb7\x48\x19\xc5\xe4\xf9\xae\xc1\x78\x8c\x12\x21\x53\x9c\x22\x94\xae\x8d\x2c\xed\x1f\x6f\x45\x39\x16\x87\x90\x0a\xbd\xea\x3b\xca\x04\xe8\xda\xd7\x25\x54\xaf\x1a\x91\x5c\xe2\xba\x99\x4c\xe1\xb3\x01\xbd\x2e\xfa\x70\xb7\xce\xe5\xca\xe5\xb0\x06\xfe\xf9\x43\xbd\x0e\x62\xdd\x74\xde\x16\x37\x9c\x4f\x96\x81\x7f\x3f\xad\x37\x08\x59\x53\x9e\xc3\x95\x96\x99\xd7\x58\x24\xe4\x89\x01\x4f\x23\x78\x6a\xaf\xda\x84\x5a\x1d\xde\xde\x2c\x9c\x42\x8a\xcf\x5e\x5d\x98\x91\x3d\xb8\x97\x63\x11\xfa\xe3\xdf\x0d\xb3\xbd\x00\x0e\x88\xfc\xc8\xbf\x7b\x58\x2c\xa3\xe9\x9d\xff\x75\x19\x4e\xa7\xd1\x62\x1a\x3d\xce\xde\x42\x97\x48\xf1\xc4\x56\xf7\x54\xdd\xc2\xa6\x03\xf2\x05\x36\x1e\x31\x20\x50\x6f\x46\x0a\x40\x33\xb1\x1a\x3d\x4b\x83\x7b\x66\xbd\x77\xc0\x36\xfc\x01\xf0\x62\x1a\xc4\xd1\xd7\xa5\x3f\x99\x44\x83\x22\xd2\x34\xd5\xbf\x88\xf6\xff\xf9\x3c\x5e\xc4\x91\x1f\xfe\x2b\xe9\x76\xdf\xf0\x2c\xb8\x5e\xc6\xf3\xdb\x69\x30\x0c\xa2\xfd\x72\x47\x28\x5f\x40\x0c\x42\xf8\xb0\x98\x46\x43\xe2\xe5\x66\x78\xbc\x9b\xf9\x22\x5e\x86\xf3\x28\x1e\x10\xd1\xfe\x36\x46\x4a\x6a\x1c\x04\x73\x3c\x0f\x82\xe9\x38\x9e\x0f\x5a\xca\x44\x0a\x51\xbc\x1f\x03\xd6\x73\x07\x3a\x70\x51\x77\xb0\xc3\x56\x36\x98\xc6\x5f\xe6\xd1\xed\x72\x36\x19\x86\x53\x03\xa7\x1b\x01\xf8\x2a\xf5\xcb\x88\xa5\x27\x32\x6e\xff\x0e\x90\xfd\x87\xc9\x2c\x5e\x86\x7e\x7c\x33\x0c\x6c\x31\x98\x8d\x14\xc5\xe7\x13\x39\x59\x56\x4c\x5b\x67\xdf\xbf\x93\xd1\x44\xbe\x0a\x2e\x69\x3a\xb1\x4f\x54\x64\xcb\x30\xb3\xbb\xe4\xc7\x8f\xb3\xb6\x7d\x98\x73\x1e\x4a\xce\x92\x8d\x47\x66\x4f\x81\xc4\x50\x83\xbd\x18\x9d\x9f\x9d\x5e\x1b\x20\xe8\x8c\x09\x8a\x4c\x8a\x7b\x30\xc6\xc6\x2c\x5e\xc8\xf3\x14\xd6\xe7\x8d\x4d\x3b\x52\x1c\x73\x2a\x21\xec\x68\x58\x9b\x55\x33\x63\x2e\xb0\x35\xb7\x64\x76\xa5\x3c\x46\x2a\x3c\x2f\xf1\xce\xab\xe1\xb6\x0d\xdf\x5c\x4d\x85\xa9\x4e\x1a\xf3\xdc\x20\xe8\x2b\xa6\xeb\x97\x53\x83\x41\xaa\xb1\xb2\xf0\xf9\x2b\xdd\x18\xe7\x00\xf9\x5a\xd3\x04\x42\xd0\x4c\xa6\xf5\x18\xf8\xe1\xc2\x69\x32\x37\x86\xae\x43\x08\x42\x54\xd7\x00\xdf\x6c\x8b\xc4\x2e\x04\xc7\x87\x78\xa7\x57\x95\x8c\xab\x7e\x7b\x9b\x26\xa8\x1b\xa9\x7f\x28\xae\x62\xec\xba\xd4\x23\x67\x87\xb5\x3f\x2b\x26\xe1\xfa\x5d\xed\xea\xc6\x45\xb1\xed\xa7\xa9\x2e\xbb\xb1\x63\x90\xe9\xf2\x0b\x01\xf4\x8d\x34\x58\x7a\xed\x9e\xc6\x3e\xe3\xd8\x6e\x96\xd6\xbb\x97\xaa\xcb\xfa\xc1\xb4\xad\xf7\x2e\xe3\x2e\x97\x71\x65\xd2\xed\x57\xdf\x8b\xbd\xbf\x47\x9b\x4a\x28\x35\x36\x01\x8f\x7a\x59\xc8\x3d\xa7\xbd\xfb\xad\xf7\xac\x60\x6b\x32\x9b\x14\x7e\xfd\x6d\xb3\x00\xbd\x66\x7b\x52\xf6\x97\x24\x94\xcd\xa5\x14\x8f\xdb\x38\x98\x28\xb7\xa0\x2e\x5a\x7d\x9b\xea\xc7\x8f\x5b\x89\xa1\xb4\x44\x99\x48\xee\x91\x78\x1c\x16\x2b\x48\xf5\x0a\x8a\x84\x2b\xab\x2a\xce\x33\x62\x19\xc8\xcd\xa4\x60\x28\x6d\xeb\x34\x62\x7e\xba\xb8\x78\x4b\xd0\xd2\xac\xad\xdd\xfa\x04\xd6\x69\x42\x95\x0a\x21\xb1\xb8\x35\xcc\x1b\x75\xab\x8b\x94\xf1\x63\x3f\xd7\xc6\xfe\x4f\xca\xdf\xa7\x59\xff\x57\x48\xd6\xb6\xc0\x3d\x55\xc0\x56\x38\x83\xa8\xd8\x2a\xd8\xc9\x52\xb6\xf6\x7f\x17\xac\xef\x82\xf5\x5d\xb0\xfe\xb1\x82\xf5\x3f\x3e\x5d\xbb\xad\x7b\xec\xcf\x1a\xb1\x7f\xdf\x00\xfc\x4f\x00\x00\x00\xff\xff\xda\x9c\xd6\x9a\x96\x17\x00\x00"),
		},
	}
	fs["/"].(*vfsgen۰DirInfo).entries = []os.FileInfo{
		fs["/agent_templates.yaml"].(os.FileInfo),
		fs["/relay_agent_template.yaml"].(os.FileInfo),
		fs["/relay_template.yaml"].(os.FileInfo),
	}

	return fs
}()

type vfsgen۰FS map[string]interface{}

func (fs vfsgen۰FS) Open(path string) (http.File, error) {
	path = pathpkg.Clean("/" + path)
	f, ok := fs[path]
	if !ok {
		return nil, &os.PathError{Op: "open", Path: path, Err: os.ErrNotExist}
	}

	switch f := f.(type) {
	case *vfsgen۰CompressedFileInfo:
		gr, err := gzip.NewReader(bytes.NewReader(f.compressedContent))
		if err != nil {
			// This should never happen because we generate the gzip bytes such that they are always valid.
			panic("unexpected error reading own gzip compressed bytes: " + err.Error())
		}
		return &vfsgen۰CompressedFile{
			vfsgen۰CompressedFileInfo: f,
			gr:                        gr,
		}, nil
	case *vfsgen۰DirInfo:
		return &vfsgen۰Dir{
			vfsgen۰DirInfo: f,
		}, nil
	default:
		// This should never happen because we generate only the above types.
		panic(fmt.Sprintf("unexpected type %T", f))
	}
}

// vfsgen۰CompressedFileInfo is a static definition of a gzip compressed file.
type vfsgen۰CompressedFileInfo struct {
	name              string
	modTime           time.Time
	compressedContent []byte
	uncompressedSize  int64
}

func (f *vfsgen۰CompressedFileInfo) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("cannot Readdir from file %s", f.name)
}
func (f *vfsgen۰CompressedFileInfo) Stat() (os.FileInfo, error) { return f, nil }

func (f *vfsgen۰CompressedFileInfo) GzipBytes() []byte {
	return f.compressedContent
}

func (f *vfsgen۰CompressedFileInfo) Name() string       { return f.name }
func (f *vfsgen۰CompressedFileInfo) Size() int64        { return f.uncompressedSize }
func (f *vfsgen۰CompressedFileInfo) Mode() os.FileMode  { return 0444 }
func (f *vfsgen۰CompressedFileInfo) ModTime() time.Time { return f.modTime }
func (f *vfsgen۰CompressedFileInfo) IsDir() bool        { return false }
func (f *vfsgen۰CompressedFileInfo) Sys() interface{}   { return nil }

// vfsgen۰CompressedFile is an opened compressedFile instance.
type vfsgen۰CompressedFile struct {
	*vfsgen۰CompressedFileInfo
	gr      *gzip.Reader
	grPos   int64 // Actual gr uncompressed position.
	seekPos int64 // Seek uncompressed position.
}

func (f *vfsgen۰CompressedFile) Read(p []byte) (n int, err error) {
	if f.grPos > f.seekPos {
		// Rewind to beginning.
		err = f.gr.Reset(bytes.NewReader(f.compressedContent))
		if err != nil {
			return 0, err
		}
		f.grPos = 0
	}
	if f.grPos < f.seekPos {
		// Fast-forward.
		_, err = io.CopyN(ioutil.Discard, f.gr, f.seekPos-f.grPos)
		if err != nil {
			return 0, err
		}
		f.grPos = f.seekPos
	}
	n, err = f.gr.Read(p)
	f.grPos += int64(n)
	f.seekPos = f.grPos
	return n, err
}
func (f *vfsgen۰CompressedFile) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekStart:
		f.seekPos = 0 + offset
	case io.SeekCurrent:
		f.seekPos += offset
	case io.SeekEnd:
		f.seekPos = f.uncompressedSize + offset
	default:
		panic(fmt.Errorf("invalid whence value: %v", whence))
	}
	return f.seekPos, nil
}
func (f *vfsgen۰CompressedFile) Close() error {
	return f.gr.Close()
}

// vfsgen۰DirInfo is a static definition of a directory.
type vfsgen۰DirInfo struct {
	name    string
	modTime time.Time
	entries []os.FileInfo
}

func (d *vfsgen۰DirInfo) Read([]byte) (int, error) {
	return 0, fmt.Errorf("cannot Read from directory %s", d.name)
}
func (d *vfsgen۰DirInfo) Close() error               { return nil }
func (d *vfsgen۰DirInfo) Stat() (os.FileInfo, error) { return d, nil }

func (d *vfsgen۰DirInfo) Name() string       { return d.name }
func (d *vfsgen۰DirInfo) Size() int64        { return 0 }
func (d *vfsgen۰DirInfo) Mode() os.FileMode  { return 0755 | os.ModeDir }
func (d *vfsgen۰DirInfo) ModTime() time.Time { return d.modTime }
func (d *vfsgen۰DirInfo) IsDir() bool        { return true }
func (d *vfsgen۰DirInfo) Sys() interface{}   { return nil }

// vfsgen۰Dir is an opened dir instance.
type vfsgen۰Dir struct {
	*vfsgen۰DirInfo
	pos int // Position within entries for Seek and Readdir.
}

func (d *vfsgen۰Dir) Seek(offset int64, whence int) (int64, error) {
	if offset == 0 && whence == io.SeekStart {
		d.pos = 0
		return 0, nil
	}
	return 0, fmt.Errorf("unsupported Seek in directory %s", d.name)
}

func (d *vfsgen۰Dir) Readdir(count int) ([]os.FileInfo, error) {
	if d.pos >= len(d.entries) && count > 0 {
		return nil, io.EOF
	}
	if count <= 0 || count > len(d.entries)-d.pos {
		count = len(d.entries) - d.pos
	}
	e := d.entries[d.pos : d.pos+count]
	d.pos += count
	return e, nil
}
