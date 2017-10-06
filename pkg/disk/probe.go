/*
 * This file is part of the Dicot project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2017 Red Hat, Inc.
 *
 */

package disk

import (
	"bytes"
	"encoding/binary"
)

type DiskFormat string

type DiskEncryption string

type DiskInfo struct {
	Format        DiskFormat
	Capacity      uint64
	BackingFile   string
	BackingFormat DiskFormat
	Encryption    DiskEncryption
}

type diskProbeInfo struct {
	Format DiskFormat

	MagicOffset int
	Magic       []byte

	ByteOrder binary.ByteOrder

	VersionOffset  int
	VersionBytes   int
	VersionNumbers []uint64

	SizeOffset     int
	SizeBytes      int
	SizeMultiplier int
}

const (
	DISK_FORMAT_AMI   = DiskFormat("ami")
	DISK_FORMAT_ARI   = DiskFormat("ari")
	DISK_FORMAT_AKI   = DiskFormat("aki")
	DISK_FORMAT_VHD   = DiskFormat("vhd")
	DISK_FORMAT_VHDX  = DiskFormat("vhdx")
	DISK_FORMAT_VMDK  = DiskFormat("vmdk")
	DISK_FORMAT_RAW   = DiskFormat("raw")
	DISK_FORMAT_QCOW2 = DiskFormat("qcow2")
	DISK_FORMAT_VDI   = DiskFormat("vdi")
	DISK_FORMAT_PLOOP = DiskFormat("ploop")
	DISK_FORMAT_ISO   = DiskFormat("iso")

	DISK_ENCRYPTION_QCOW = DiskEncryption("qcow")
	DISK_ENCRYPTION_LUKS = DiskEncryption("luks")

	qcowxHdrVersion           = 4
	qcowxHdrBackingFileOffset = qcowxHdrVersion + 4
	qcowxHdrBackingFileSize   = qcowxHdrBackingFileOffset + 8
	qcowxHdrImageSize         = qcowxHdrBackingFileSize + 4 + 4
)

var (
	diskProbeTable = []diskProbeInfo{
		diskProbeInfo{
			DISK_FORMAT_QCOW2,
			0, []byte("QFI"), binary.BigEndian,
			4, 4, []uint64{2, 3},

			qcowxHdrImageSize, 8, 1,
		},

		diskProbeInfo{
			DISK_FORMAT_ISO,
			32769, []byte("CD001"), binary.LittleEndian,
			-1, 0, []uint64{},
			-1, 0, 0,
		},

		diskProbeInfo{
			DISK_FORMAT_AKI,
			512, []byte("HdrS"), binary.LittleEndian,
			-1, 0, []uint64{},
			-1, 0, 0,
		},

		diskProbeInfo{
			DISK_FORMAT_RAW,
			/* MBR signature */
			0x1fe, []byte{0xAA, 0x55}, binary.LittleEndian,
			-1, 0, []uint64{},
			-1, 0, 0,
		},
	}
)

func matchesMagic(header []byte, probe diskProbeInfo) bool {
	if len(header) < (probe.MagicOffset + len(probe.Magic)) {
		return false
	}

	if bytes.Compare(probe.Magic, header[probe.MagicOffset:probe.MagicOffset+len(probe.Magic)]) != 0 {
		return false
	}

	return true
}

func matchesVersion(header []byte, probe diskProbeInfo) bool {
	if probe.VersionOffset == -1 { /* non-versioned file */
		return true
	}

	if len(header) < (probe.VersionOffset + probe.VersionBytes) {
		return false
	}

	var val uint64
	switch probe.VersionBytes {
	case 2:
		val = uint64(probe.ByteOrder.Uint16(header[probe.VersionOffset : probe.VersionOffset+2]))
	case 4:
		val = uint64(probe.ByteOrder.Uint32(header[probe.VersionOffset : probe.VersionOffset+4]))
	case 8:
		val = probe.ByteOrder.Uint64(header[probe.VersionOffset : probe.VersionOffset+8])
	default:
		return false
	}

	for _, ver := range probe.VersionNumbers {
		if val == ver {
			return true
		}
	}
	return false
}

func getCapacity(header []byte, probe diskProbeInfo) (bool, uint64) {
	if len(header) < (probe.SizeOffset + probe.SizeBytes) {
		return false, 0
	}

	switch probe.SizeBytes {
	case 2:
		return true, uint64(probe.ByteOrder.Uint16(header[probe.SizeOffset : probe.SizeOffset+2]))
	case 4:
		return true, uint64(probe.ByteOrder.Uint32(header[probe.SizeOffset : probe.SizeOffset+4]))
	case 8:
		return true, probe.ByteOrder.Uint64(header[probe.SizeOffset : probe.SizeOffset+8])
	}

	return false, 0
}

func probeFormat(header []byte, probe diskProbeInfo) (info *DiskInfo) {
	if !matchesMagic(header, probe) || !matchesVersion(header, probe) {
		return nil
	}

	ok, capacity := getCapacity(header, probe)
	if !ok {
		return nil
	}

	return &DiskInfo{
		probe.Format,
		capacity,
		"",
		"",
		"",
	}
}

func ProbeFormat(header []byte) *DiskInfo {
	for _, probe := range diskProbeTable {
		info := probeFormat(header, probe)
		if info != nil {
			return info
		}
	}

	return nil
}
