package api

import (
	"bytes"
	"crypto/sha256"
	"fmt"

	cid "github.com/drausin/libri/libri/common/id"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

const (
	// ECPubKeyLength is the length of a 256-bit ECDSA public key point serialized
	// (uncompressed) to a byte string.
	ECPubKeyLength = 65

	// DocumentKeyLength is the byte length a document's key.
	DocumentKeyLength = cid.Length

	// AESKeyLength is the byte length of an AES-256 encryption key.
	AESKeyLength = 32

	// PageIVSeedLength is the byte length of the Page block cipher initialization vector (IV)
	// seed.
	PageIVSeedLength = 32

	// PageHMACKeyLength is the byte length of the Page HMAC-256 key.
	PageHMACKeyLength = 32

	// MetadataIVLength is the byte length of the metadata block cipher initialization vector.
	MetadataIVLength = 12

	// EncryptionKeysLength is the total byte length of all the keys used to encrypt an Entry.
	EncryptionKeysLength = AESKeyLength +
		PageIVSeedLength +
		PageHMACKeyLength +
		MetadataIVLength

	// HMAC256Length is the byte length of an HMAC-256.
	HMAC256Length = sha256.Size
)

// GetKey calculates the key from the has of the proto.Message.
func GetKey(value proto.Message) (cid.ID, error) {
	valueBytes, err := proto.Marshal(value)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(valueBytes)
	return cid.FromBytes(hash[:]), nil
}

// ValidateDocument checks that all fields of a Document are populated and have the expected
// lengths.
func ValidateDocument(d *Document) error {
	if d == nil {
		return errors.New("Document may not be nil")
	}
	switch x := d.Contents.(type) {
	case *Document_Envelope:
		return ValidateEnvelope(x.Envelope)
	case *Document_Entry:
		return ValidateEntry(x.Entry)
	case *Document_Page:
		return ValidatePage(x.Page)
	default:
		return errors.New("unknown document type")
	}
}

// ValidateEnvelope checks that all fields of an Envelope are populated and have the expected
// lengths.
func ValidateEnvelope(e *Envelope) error {
	if e == nil {
		return errors.New("Envelope may not be nil")
	}
	if err := ValidatePublicKey(e.AuthorPublicKey); err != nil {
		return err
	}
	if err := ValidatePublicKey(e.ReaderPublicKey); err != nil {
		return err
	}
	if err := ValidateBytes(e.EntryKey, DocumentKeyLength, "EntryPublicKey"); err != nil {
		return err
	}
	if err := ValidateBytes(e.EncryptionKeysCiphertext, EncryptionKeysLength,
		"EncryptionKeysLength"); err != nil {
		return err
	}

	return nil
}

// ValidateEntry checks that all fields of an Entry are populated and have the expected byte
// lengths.
func ValidateEntry(e *Entry) error {
	if e == nil {
		return errors.New("Entry may not be nil")
	}
	if err := ValidatePublicKey(e.AuthorPublicKey); err != nil {
		return err
	}
	if e.CreatedTime == 0 {
		return errors.New("CreateTime must be populated")
	}
	if err := ValidateHMAC256(e.MetadataCiphertextMac); err != nil {
		return err
	}
	if err := validateNotEmpty(e.MetadataCiphertext, "MetadataCiphertext"); err != nil {
		return err
	}
	if err := ValidateHMAC256(e.ContentsCiphertextMac); err != nil {
		return err
	}
	if e.ContentsCiphertextSize == 0 {
		return errors.New("ContentsCiphertextSize must be greateer than zero")
	}
	if err := validateEntryContents(e); err != nil {
		return err
	}
	return nil
}

func validateEntryContents(e *Entry) error {
	switch x := e.Contents.(type) {
	case *Entry_Page:
		if !bytes.Equal(e.AuthorPublicKey, x.Page.AuthorPublicKey) {
			return errors.New("Page author public key must be the same as Entry's")
		}
		return ValidatePage(x.Page)
	case *Entry_PageKeys:
		return ValidatePageKeys(x.PageKeys)
	default:
		return errors.New("unknown Entry.Contents type")
	}
}

// ValidatePage checks that all fields of a Page are populated and have the expected lengths.
func ValidatePage(p *Page) error {
	if p == nil {
		return errors.New("Page may not be nil")
	}
	if err := ValidatePublicKey(p.AuthorPublicKey); err != nil {
		return err
	}
	// nothing to check for index, since it's zero value is legitimate
	if err := ValidateHMAC256(p.CiphertextMac); err != nil {
		return err
	}
	if err := validateNotEmpty(p.Ciphertext, "Ciphertext"); err != nil {
		return err
	}
	return nil
}

// ValidatePageKeys checks that all fields of a PageKeys are populated and have the expected
// lengths.
func ValidatePageKeys(pk *PageKeys) error {
	if pk == nil {
		return errors.New("PageKeys may not be nil")
	}
	if pk.Keys == nil {
		return errors.New("PageKeys.Keys may not be nil")
	}
	if len(pk.Keys) == 0 {
		return errors.New("PageKeys.Keys must have length > 0")
	}
	for i, k := range pk.Keys {
		if err := validateNotEmpty(k, fmt.Sprintf("key %d", i)); err != nil {
			return err
		}
	}
	return nil
}

// ValidatePublicKey checks that a value can be a 256-bit elliptic curve public key.
func ValidatePublicKey(value []byte) error {
	return ValidateBytes(value, ECPubKeyLength, "PublicKey")
}

// ValidateAESKey checks the a value can be a 256-bit AES key.
func ValidateAESKey(value []byte) error {
	return ValidateBytes(value, AESKeyLength, "AESKey")
}

// ValidatePageHMACKey checks that a value can be an HMAC-256 key.
func ValidatePageHMACKey(value []byte) error {
	return ValidateBytes(value, PageHMACKeyLength, "PageHMACKey")
}

// ValidatePageIVSeed checks that a value can be a 256-bit initialization vector seed.
func ValidatePageIVSeed(value []byte) error {
	return ValidateBytes(value, PageIVSeedLength, "PageIVSeed")
}

// ValidateMetadataIV checks that a value can be a 12-byte GCM initialization vector.
func ValidateMetadataIV(value []byte) error {
	return ValidateBytes(value, MetadataIVLength, "MetadataIV")
}

// ValidateHMAC256 checks that a value can be an HMAC-256.
func ValidateHMAC256(value []byte) error {
	return ValidateBytes(value, HMAC256Length, "HMAC256")
}

// ValidateBytes returns whether the byte slice is not empty and has an expected length.
func ValidateBytes(value []byte, expectedLen int, name string) error {
	if err := validateNotEmpty(value, name); err != nil {
		return err
	}
	if len(value) != expectedLen {
		return fmt.Errorf("%s must have length %d, found length %d", name, expectedLen,
			len(value))
	}
	return nil
}

func validateNotEmpty(value []byte, name string) error {
	if value == nil {
		return fmt.Errorf("%s may not be nil", name)
	}
	if len(value) == 0 {
		return fmt.Errorf("%s must have length > 0", name)
	}
	if allZeros(value) {
		return fmt.Errorf("%s may not be all zeros", name)
	}
	return nil
}

func allZeros(value []byte) bool {
	for _, v := range value {
		if v != 0 {
			return false
		}
	}
	return true
}