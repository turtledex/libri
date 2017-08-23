// Code generated by protoc-gen-go.
// source: librarian/api/documents.proto
// DO NOT EDIT!

/*
Package api is a generated protocol buffer package.

It is generated from these files:
	librarian/api/documents.proto
	librarian/api/librarian.proto

It has these top-level messages:
	Document
	Envelope
	Entry
	Metadata
	PageKeys
	Page
	RequestMetadata
	ResponseMetadata
	IntroduceRequest
	IntroduceResponse
	FindRequest
	FindResponse
	VerifyRequest
	VerifyResponse
	PeerAddress
	StoreRequest
	StoreResponse
	GetRequest
	GetResponse
	PutRequest
	PutResponse
	SubscribeRequest
	SubscribeResponse
	Publication
	Subscription
	BloomFilter
*/
package api

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// Document contains either an Envelope, Entry, or Page message.
type Document struct {
	// Types that are valid to be assigned to Contents:
	//	*Document_Envelope
	//	*Document_Entry
	//	*Document_Page
	Contents isDocument_Contents `protobuf_oneof:"contents"`
}

func (m *Document) Reset()                    { *m = Document{} }
func (m *Document) String() string            { return proto.CompactTextString(m) }
func (*Document) ProtoMessage()               {}
func (*Document) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type isDocument_Contents interface {
	isDocument_Contents()
}

type Document_Envelope struct {
	Envelope *Envelope `protobuf:"bytes,1,opt,name=envelope,oneof"`
}
type Document_Entry struct {
	Entry *Entry `protobuf:"bytes,2,opt,name=entry,oneof"`
}
type Document_Page struct {
	Page *Page `protobuf:"bytes,3,opt,name=page,oneof"`
}

func (*Document_Envelope) isDocument_Contents() {}
func (*Document_Entry) isDocument_Contents()    {}
func (*Document_Page) isDocument_Contents()     {}

func (m *Document) GetContents() isDocument_Contents {
	if m != nil {
		return m.Contents
	}
	return nil
}

func (m *Document) GetEnvelope() *Envelope {
	if x, ok := m.GetContents().(*Document_Envelope); ok {
		return x.Envelope
	}
	return nil
}

func (m *Document) GetEntry() *Entry {
	if x, ok := m.GetContents().(*Document_Entry); ok {
		return x.Entry
	}
	return nil
}

func (m *Document) GetPage() *Page {
	if x, ok := m.GetContents().(*Document_Page); ok {
		return x.Page
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Document) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Document_OneofMarshaler, _Document_OneofUnmarshaler, _Document_OneofSizer, []interface{}{
		(*Document_Envelope)(nil),
		(*Document_Entry)(nil),
		(*Document_Page)(nil),
	}
}

func _Document_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Document)
	// contents
	switch x := m.Contents.(type) {
	case *Document_Envelope:
		b.EncodeVarint(1<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Envelope); err != nil {
			return err
		}
	case *Document_Entry:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Entry); err != nil {
			return err
		}
	case *Document_Page:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Page); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Document.Contents has unexpected type %T", x)
	}
	return nil
}

func _Document_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Document)
	switch tag {
	case 1: // contents.envelope
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Envelope)
		err := b.DecodeMessage(msg)
		m.Contents = &Document_Envelope{msg}
		return true, err
	case 2: // contents.entry
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Entry)
		err := b.DecodeMessage(msg)
		m.Contents = &Document_Entry{msg}
		return true, err
	case 3: // contents.page
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Page)
		err := b.DecodeMessage(msg)
		m.Contents = &Document_Page{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Document_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Document)
	// contents
	switch x := m.Contents.(type) {
	case *Document_Envelope:
		s := proto.Size(x.Envelope)
		n += proto.SizeVarint(1<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Document_Entry:
		s := proto.Size(x.Entry)
		n += proto.SizeVarint(2<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Document_Page:
		s := proto.Size(x.Page)
		n += proto.SizeVarint(3<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Envelope defines the public keys an author uses to share an entry encryption key with a
// particular reader. The shared ECDH secret is used with a key derivation function to generate the
// key encryption key (KEK), which contains 2 sub-keys:
// 1) 32-byte AES-256 key, used for encrypting/decrypting the entry encryption key (EEK)
// 2) 12-byte EEK block cipher initialization vector
// 3) 32-byte HMAC-256 key, used for EEK MAC
//
// This AES-256 KEK is then used to decrypt the EEK ciphertext, which contains 4 sub-keys:
// 1) 32-byte AES-256 key, used to enrypt Pages and Entry metadata
// 2) 32-byte Page initialization vector (IV) seed
// 3) 32-byte HMAC-256 key
// 4) 12-byte metadata block cipher initialization vector
type Envelope struct {
	// 32-byte key of the Entry whose encryption keys are being sent
	EntryKey []byte `protobuf:"bytes,1,opt,name=entry_key,json=entryKey,proto3" json:"entry_key,omitempty"`
	// ECDH public key of the entry author/sender
	AuthorPublicKey []byte `protobuf:"bytes,2,opt,name=author_public_key,json=authorPublicKey,proto3" json:"author_public_key,omitempty"`
	// ECDH public key of the entry reader/recipient
	ReaderPublicKey []byte `protobuf:"bytes,3,opt,name=reader_public_key,json=readerPublicKey,proto3" json:"reader_public_key,omitempty"`
	// ciphertext of 108-byte entry encryption key (EEK), encrypted with a KEK from the shared
	// ECDH shared secret
	EekCiphertext []byte `protobuf:"bytes,4,opt,name=eek_ciphertext,json=eekCiphertext,proto3" json:"eek_ciphertext,omitempty"`
	// 32-byte MAC of the EEK
	EekCiphertextMac []byte `protobuf:"bytes,5,opt,name=eek_ciphertext_mac,json=eekCiphertextMac,proto3" json:"eek_ciphertext_mac,omitempty"`
}

func (m *Envelope) Reset()                    { *m = Envelope{} }
func (m *Envelope) String() string            { return proto.CompactTextString(m) }
func (*Envelope) ProtoMessage()               {}
func (*Envelope) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Envelope) GetEntryKey() []byte {
	if m != nil {
		return m.EntryKey
	}
	return nil
}

func (m *Envelope) GetAuthorPublicKey() []byte {
	if m != nil {
		return m.AuthorPublicKey
	}
	return nil
}

func (m *Envelope) GetReaderPublicKey() []byte {
	if m != nil {
		return m.ReaderPublicKey
	}
	return nil
}

func (m *Envelope) GetEekCiphertext() []byte {
	if m != nil {
		return m.EekCiphertext
	}
	return nil
}

func (m *Envelope) GetEekCiphertextMac() []byte {
	if m != nil {
		return m.EekCiphertextMac
	}
	return nil
}

// Entry is the main unit of storage in the Libri network.
type Entry struct {
	// ECDSA public key of the entry author
	AuthorPublicKey []byte `protobuf:"bytes,1,opt,name=author_public_key,json=authorPublicKey,proto3" json:"author_public_key,omitempty"`
	// contents of the entry, either a single Page or a list of page keys
	//
	// Types that are valid to be assigned to Contents:
	//	*Entry_Page
	//	*Entry_PageKeys
	Contents isEntry_Contents `protobuf_oneof:"contents"`
	// created epoch time (seconds since 1970-01-01)
	CreatedTime int64 `protobuf:"varint,4,opt,name=created_time,json=createdTime" json:"created_time,omitempty"`
	// ciphertext of marshalled Metadata message properties
	MetadataCiphertext []byte `protobuf:"bytes,5,opt,name=metadata_ciphertext,json=metadataCiphertext,proto3" json:"metadata_ciphertext,omitempty"`
	// 32-byte MAC of metatadata ciphertext, encrypted with the 32-byte Entry AES-256 key and
	// 12-byte metadata block cipher IV
	MetadataCiphertextMac []byte `protobuf:"bytes,6,opt,name=metadata_ciphertext_mac,json=metadataCiphertextMac,proto3" json:"metadata_ciphertext_mac,omitempty"`
}

func (m *Entry) Reset()                    { *m = Entry{} }
func (m *Entry) String() string            { return proto.CompactTextString(m) }
func (*Entry) ProtoMessage()               {}
func (*Entry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type isEntry_Contents interface {
	isEntry_Contents()
}

type Entry_Page struct {
	Page *Page `protobuf:"bytes,2,opt,name=page,oneof"`
}
type Entry_PageKeys struct {
	PageKeys *PageKeys `protobuf:"bytes,3,opt,name=page_keys,json=pageKeys,oneof"`
}

func (*Entry_Page) isEntry_Contents()     {}
func (*Entry_PageKeys) isEntry_Contents() {}

func (m *Entry) GetContents() isEntry_Contents {
	if m != nil {
		return m.Contents
	}
	return nil
}

func (m *Entry) GetAuthorPublicKey() []byte {
	if m != nil {
		return m.AuthorPublicKey
	}
	return nil
}

func (m *Entry) GetPage() *Page {
	if x, ok := m.GetContents().(*Entry_Page); ok {
		return x.Page
	}
	return nil
}

func (m *Entry) GetPageKeys() *PageKeys {
	if x, ok := m.GetContents().(*Entry_PageKeys); ok {
		return x.PageKeys
	}
	return nil
}

func (m *Entry) GetCreatedTime() int64 {
	if m != nil {
		return m.CreatedTime
	}
	return 0
}

func (m *Entry) GetMetadataCiphertext() []byte {
	if m != nil {
		return m.MetadataCiphertext
	}
	return nil
}

func (m *Entry) GetMetadataCiphertextMac() []byte {
	if m != nil {
		return m.MetadataCiphertextMac
	}
	return nil
}

// XXX_OneofFuncs is for the internal use of the proto package.
func (*Entry) XXX_OneofFuncs() (func(msg proto.Message, b *proto.Buffer) error, func(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error), func(msg proto.Message) (n int), []interface{}) {
	return _Entry_OneofMarshaler, _Entry_OneofUnmarshaler, _Entry_OneofSizer, []interface{}{
		(*Entry_Page)(nil),
		(*Entry_PageKeys)(nil),
	}
}

func _Entry_OneofMarshaler(msg proto.Message, b *proto.Buffer) error {
	m := msg.(*Entry)
	// contents
	switch x := m.Contents.(type) {
	case *Entry_Page:
		b.EncodeVarint(2<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.Page); err != nil {
			return err
		}
	case *Entry_PageKeys:
		b.EncodeVarint(3<<3 | proto.WireBytes)
		if err := b.EncodeMessage(x.PageKeys); err != nil {
			return err
		}
	case nil:
	default:
		return fmt.Errorf("Entry.Contents has unexpected type %T", x)
	}
	return nil
}

func _Entry_OneofUnmarshaler(msg proto.Message, tag, wire int, b *proto.Buffer) (bool, error) {
	m := msg.(*Entry)
	switch tag {
	case 2: // contents.page
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(Page)
		err := b.DecodeMessage(msg)
		m.Contents = &Entry_Page{msg}
		return true, err
	case 3: // contents.page_keys
		if wire != proto.WireBytes {
			return true, proto.ErrInternalBadWireType
		}
		msg := new(PageKeys)
		err := b.DecodeMessage(msg)
		m.Contents = &Entry_PageKeys{msg}
		return true, err
	default:
		return false, nil
	}
}

func _Entry_OneofSizer(msg proto.Message) (n int) {
	m := msg.(*Entry)
	// contents
	switch x := m.Contents.(type) {
	case *Entry_Page:
		s := proto.Size(x.Page)
		n += proto.SizeVarint(2<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case *Entry_PageKeys:
		s := proto.Size(x.PageKeys)
		n += proto.SizeVarint(3<<3 | proto.WireBytes)
		n += proto.SizeVarint(uint64(s))
		n += s
	case nil:
	default:
		panic(fmt.Sprintf("proto: unexpected type %T in oneof", x))
	}
	return n
}

// Metadata is a map of (property, value) combinations.
type Metadata struct {
	Properties map[string][]byte `protobuf:"bytes,1,rep,name=properties" json:"properties,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (m *Metadata) Reset()                    { *m = Metadata{} }
func (m *Metadata) String() string            { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()               {}
func (*Metadata) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Metadata) GetProperties() map[string][]byte {
	if m != nil {
		return m.Properties
	}
	return nil
}

// PageKeys is an ordered list of keys to Page documents that comprise an Entry document.
type PageKeys struct {
	Keys [][]byte `protobuf:"bytes,1,rep,name=keys,proto3" json:"keys,omitempty"`
}

func (m *PageKeys) Reset()                    { *m = PageKeys{} }
func (m *PageKeys) String() string            { return proto.CompactTextString(m) }
func (*PageKeys) ProtoMessage()               {}
func (*PageKeys) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *PageKeys) GetKeys() [][]byte {
	if m != nil {
		return m.Keys
	}
	return nil
}

// Page is a portion (possibly all) of an Entry document.
type Page struct {
	// ECDSA public key of the entry author
	AuthorPublicKey []byte `protobuf:"bytes,1,opt,name=author_public_key,json=authorPublicKey,proto3" json:"author_public_key,omitempty"`
	// index of Page within Entry contents
	Index uint32 `protobuf:"varint,2,opt,name=index" json:"index,omitempty"`
	// ciphertext of Page contents, encrypted using the 32-byte AES-256 key with the block cipher
	// initialized by the first 12 bytes of HMAC-256(IV seed, page index)
	Ciphertext []byte `protobuf:"bytes,3,opt,name=ciphertext,proto3" json:"ciphertext,omitempty"`
	// 32-byte MAC of ciphertext using the 32-byte Page ciphertext HMAC-256 key
	CiphertextMac []byte `protobuf:"bytes,4,opt,name=ciphertext_mac,json=ciphertextMac,proto3" json:"ciphertext_mac,omitempty"`
}

func (m *Page) Reset()                    { *m = Page{} }
func (m *Page) String() string            { return proto.CompactTextString(m) }
func (*Page) ProtoMessage()               {}
func (*Page) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Page) GetAuthorPublicKey() []byte {
	if m != nil {
		return m.AuthorPublicKey
	}
	return nil
}

func (m *Page) GetIndex() uint32 {
	if m != nil {
		return m.Index
	}
	return 0
}

func (m *Page) GetCiphertext() []byte {
	if m != nil {
		return m.Ciphertext
	}
	return nil
}

func (m *Page) GetCiphertextMac() []byte {
	if m != nil {
		return m.CiphertextMac
	}
	return nil
}

func init() {
	proto.RegisterType((*Document)(nil), "api.Document")
	proto.RegisterType((*Envelope)(nil), "api.Envelope")
	proto.RegisterType((*Entry)(nil), "api.Entry")
	proto.RegisterType((*Metadata)(nil), "api.Metadata")
	proto.RegisterType((*PageKeys)(nil), "api.PageKeys")
	proto.RegisterType((*Page)(nil), "api.Page")
}

func init() { proto.RegisterFile("librarian/api/documents.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 489 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x94, 0x93, 0xdf, 0x6b, 0xdb, 0x30,
	0x10, 0xc7, 0x6b, 0x3b, 0x2e, 0xce, 0x25, 0x59, 0x3b, 0xad, 0x63, 0x66, 0xa3, 0x5d, 0x67, 0x18,
	0x94, 0xad, 0x24, 0xd0, 0xc1, 0x18, 0x83, 0xbe, 0xec, 0x07, 0x14, 0x4a, 0x20, 0x98, 0xbd, 0x1b,
	0x45, 0x3e, 0x5a, 0x91, 0xd8, 0x16, 0x8a, 0x52, 0xea, 0xff, 0x60, 0x6f, 0x7b, 0xdf, 0xdf, 0xb5,
	0x3f, 0x68, 0xe8, 0x2c, 0x67, 0x4e, 0x97, 0x3d, 0xec, 0xc9, 0xd2, 0x7d, 0x3f, 0x27, 0x9d, 0xbe,
	0x77, 0x86, 0xe3, 0xa5, 0x9c, 0x6b, 0xae, 0x25, 0x2f, 0x27, 0x5c, 0xc9, 0x49, 0x5e, 0x89, 0x75,
	0x81, 0xa5, 0x59, 0x8d, 0x95, 0xae, 0x4c, 0xc5, 0x02, 0xae, 0x64, 0xf2, 0xdd, 0x83, 0xe8, 0x8b,
	0x13, 0xd8, 0x5b, 0x88, 0xb0, 0xbc, 0xc3, 0x65, 0xa5, 0x30, 0xf6, 0x4e, 0xbd, 0xb3, 0xc1, 0xc5,
	0x68, 0xcc, 0x95, 0x1c, 0x7f, 0x75, 0xc1, 0xab, 0xbd, 0x74, 0x03, 0xb0, 0x04, 0x42, 0x2c, 0x8d,
	0xae, 0x63, 0x9f, 0x48, 0x70, 0xa4, 0xd1, 0xf5, 0xd5, 0x5e, 0xda, 0x48, 0xec, 0x25, 0xf4, 0x14,
	0xbf, 0xc1, 0x38, 0x20, 0xa4, 0x4f, 0xc8, 0x8c, 0xdf, 0xd8, 0x83, 0x48, 0xf8, 0x04, 0x10, 0x89,
	0xaa, 0x34, 0xb6, 0xaa, 0xe4, 0x97, 0x07, 0x51, 0x7b, 0x13, 0x7b, 0x01, 0x7d, 0x3a, 0x22, 0x5b,
	0x60, 0x4d, 0xb5, 0x0c, 0xed, 0xd5, 0x46, 0xd7, 0xd7, 0x58, 0xb3, 0x37, 0xf0, 0x98, 0xaf, 0xcd,
	0x6d, 0xa5, 0x33, 0xb5, 0x9e, 0x2f, 0xa5, 0x20, 0xc8, 0x27, 0xe8, 0xa0, 0x11, 0x66, 0x14, 0x77,
	0xac, 0x46, 0x9e, 0xe3, 0x16, 0x1b, 0x34, 0x6c, 0x23, 0xfc, 0x61, 0x5f, 0xc3, 0x23, 0xc4, 0x45,
	0x26, 0xa4, 0xba, 0x45, 0x6d, 0xf0, 0xde, 0xc4, 0x3d, 0x02, 0x47, 0x88, 0x8b, 0xcf, 0x9b, 0x20,
	0x3b, 0x07, 0xb6, 0x8d, 0x65, 0x05, 0x17, 0x71, 0x48, 0xe8, 0xe1, 0x16, 0x3a, 0xe5, 0x22, 0xf9,
	0xe9, 0x43, 0x48, 0xb6, 0xec, 0x2e, 0xdb, 0xdb, 0x5d, 0x76, 0xeb, 0x9c, 0xff, 0x0f, 0xe7, 0xd8,
	0x39, 0xf4, 0xed, 0xd7, 0x9e, 0xb1, 0x72, 0xfe, 0x8e, 0x36, 0xd4, 0x35, 0xd6, 0x2b, 0xdb, 0x2c,
	0xe5, 0xd6, 0xec, 0x15, 0x0c, 0x85, 0x46, 0x6e, 0x30, 0xcf, 0x8c, 0x2c, 0x90, 0xde, 0x15, 0xa4,
	0x03, 0x17, 0xfb, 0x26, 0x0b, 0x64, 0x13, 0x78, 0x52, 0xa0, 0xe1, 0x39, 0x37, 0xbc, 0xeb, 0x40,
	0xf3, 0x2c, 0xd6, 0x4a, 0x1d, 0x1b, 0xde, 0xc3, 0xb3, 0x1d, 0x09, 0xe4, 0xc5, 0x3e, 0x25, 0x3d,
	0xfd, 0x3b, 0x69, 0xca, 0xc5, 0x56, 0xcf, 0xed, 0xf8, 0x4d, 0x1d, 0xc5, 0x2e, 0x01, 0x94, 0xae,
	0x14, 0x6a, 0x23, 0x71, 0x15, 0x7b, 0xa7, 0xc1, 0xd9, 0xe0, 0xe2, 0x98, 0xde, 0xd4, 0x22, 0xe3,
	0xd9, 0x46, 0x27, 0x4b, 0xd3, 0x4e, 0xc2, 0xf3, 0x4b, 0x38, 0x78, 0x20, 0xb3, 0x43, 0x08, 0x5a,
	0x8f, 0xfb, 0xa9, 0x5d, 0xb2, 0x23, 0x08, 0xef, 0xf8, 0x72, 0x8d, 0x6e, 0x5c, 0x9a, 0xcd, 0x47,
	0xff, 0x83, 0x97, 0x9c, 0x40, 0xd4, 0x5a, 0xc7, 0x18, 0xf4, 0xc8, 0x57, 0x5b, 0xc3, 0x30, 0xa5,
	0x75, 0xf2, 0xc3, 0x83, 0x9e, 0x05, 0xfe, 0xab, 0x8d, 0x47, 0x10, 0xca, 0x32, 0xc7, 0x7b, 0xba,
	0x6e, 0x94, 0x36, 0x1b, 0x76, 0x02, 0xd0, 0x71, 0xb8, 0x19, 0xc6, 0x4e, 0xc4, 0xce, 0xe1, 0x03,
	0x43, 0xdd, 0x1c, 0x8a, 0xae, 0x91, 0xf3, 0x7d, 0xfa, 0x8f, 0xdf, 0xfd, 0x0e, 0x00, 0x00, 0xff,
	0xff, 0x8c, 0xae, 0x7a, 0x14, 0xe8, 0x03, 0x00, 0x00,
}
