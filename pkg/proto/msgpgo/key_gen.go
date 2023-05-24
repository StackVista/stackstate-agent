package msgpgo

// Code generated by github.com/tinylib/msgp DO NOT EDIT.

import (
	"github.com/tinylib/msgp/msgp"
)

// DecodeMsg implements msgp.Decodable
func (z *RemoteConfigKey) DecodeMsg(dc *msgp.Reader) (err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, err = dc.ReadMapHeader()
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, err = dc.ReadMapKeyPtr()
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "key":
			z.AppKey, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "AppKey")
				return
			}
		case "org":
			z.OrgID, err = dc.ReadInt64()
			if err != nil {
				err = msgp.WrapError(err, "OrgID")
				return
			}
		case "dc":
			z.Datacenter, err = dc.ReadString()
			if err != nil {
				err = msgp.WrapError(err, "Datacenter")
				return
			}
		default:
			err = dc.Skip()
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	return
}

// EncodeMsg implements msgp.Encodable
func (z RemoteConfigKey) EncodeMsg(en *msgp.Writer) (err error) {
	// map header, size 3
	// write "key"
	err = en.Append(0x83, 0xa3, 0x6b, 0x65, 0x79)
	if err != nil {
		return
	}
	err = en.WriteString(z.AppKey)
	if err != nil {
		err = msgp.WrapError(err, "AppKey")
		return
	}
	// write "org"
	err = en.Append(0xa3, 0x6f, 0x72, 0x67)
	if err != nil {
		return
	}
	err = en.WriteInt64(z.OrgID)
	if err != nil {
		err = msgp.WrapError(err, "OrgID")
		return
	}
	// write "dc"
	err = en.Append(0xa2, 0x64, 0x63)
	if err != nil {
		return
	}
	err = en.WriteString(z.Datacenter)
	if err != nil {
		err = msgp.WrapError(err, "Datacenter")
		return
	}
	return
}

// MarshalMsg implements msgp.Marshaler
func (z RemoteConfigKey) MarshalMsg(b []byte) (o []byte, err error) {
	o = msgp.Require(b, z.Msgsize())
	// map header, size 3
	// string "key"
	o = append(o, 0x83, 0xa3, 0x6b, 0x65, 0x79)
	o = msgp.AppendString(o, z.AppKey)
	// string "org"
	o = append(o, 0xa3, 0x6f, 0x72, 0x67)
	o = msgp.AppendInt64(o, z.OrgID)
	// string "dc"
	o = append(o, 0xa2, 0x64, 0x63)
	o = msgp.AppendString(o, z.Datacenter)
	return
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *RemoteConfigKey) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var field []byte
	_ = field
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadMapHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	for zb0001 > 0 {
		zb0001--
		field, bts, err = msgp.ReadMapKeyZC(bts)
		if err != nil {
			err = msgp.WrapError(err)
			return
		}
		switch msgp.UnsafeString(field) {
		case "key":
			z.AppKey, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "AppKey")
				return
			}
		case "org":
			z.OrgID, bts, err = msgp.ReadInt64Bytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "OrgID")
				return
			}
		case "dc":
			z.Datacenter, bts, err = msgp.ReadStringBytes(bts)
			if err != nil {
				err = msgp.WrapError(err, "Datacenter")
				return
			}
		default:
			bts, err = msgp.Skip(bts)
			if err != nil {
				err = msgp.WrapError(err)
				return
			}
		}
	}
	o = bts
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z RemoteConfigKey) Msgsize() (s int) {
	s = 1 + 4 + msgp.StringPrefixSize + len(z.AppKey) + 4 + msgp.Int64Size + 3 + msgp.StringPrefixSize + len(z.Datacenter)
	return
}