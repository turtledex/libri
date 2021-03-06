package api

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"

	cerrors "github.com/drausin/libri/libri/common/errors"
	"github.com/drausin/libri/libri/common/id"
	"github.com/golang/protobuf/proto"
)

// field lengths
const (
	// ECPubKeyLength is the length of a 256-bit ECDSA public key point serialized
	// (compressed) to a byte string.
	ECPubKeyLength = 33

	// DocumentKeyLength is the byte length a document's key.
	DocumentKeyLength = id.Length

	// AESKeyLength is the byte length of an AES-256 encryption key.
	AESKeyLength = 32

	// PageIVSeedLength is the byte length of the Page block cipher initialization vector (IV)
	// seed.
	PageIVSeedLength = 32

	// HMACKeyLength is the byte length of the Page HMAC-256 key.
	HMACKeyLength = 32

	// BlockCipherIVLength is the byte length of a block cipher initialization vector.
	BlockCipherIVLength = 12

	// KEKLength is the total byte length of the key encryption keys.
	KEKLength = AESKeyLength + BlockCipherIVLength + HMACKeyLength

	// EEKLength is the total byte length of the entry encryption keys.
	EEKLength = AESKeyLength +
		PageIVSeedLength +
		HMACKeyLength +
		BlockCipherIVLength

	// EEKCiphertextLength is the length of the EEK ciphertext, which includes 16 bytes of
	// encryption info.
	EEKCiphertextLength = EEKLength + 16

	// HMAC256Length is the byte length of an HMAC-256.
	HMAC256Length = sha256.Size
)

var (
	// ErrUnexpectedDocumentType indicates when a document type is not expected (e.g., a Page
	// when expecting an Entry).
	ErrUnexpectedDocumentType = errors.New("unexpected document type")

	// ErrUnknownDocumentType indicates when a document type is not known (usually, this
	// error should never actually be thrown).
	ErrUnknownDocumentType = errors.New("unknown document type")

	// ErrUnexpectedKey indicates when a key does not match the expected key (from GetKey)
	// for a given value.
	ErrUnexpectedKey = errors.New("unexpected key for value")

	// ErrMissingDocument indicates when a document is unexpectedly missing.
	ErrMissingDocument = errors.New("missing document")

	// ErrMissingEnvelope indicates when an envelope is unexpectedly missing.
	ErrMissingEnvelope = errors.New("missing envelope")

	// ErrMissingEntry indicates when an entry is unexpectedly missing.
	ErrMissingEntry = errors.New("missing entry")

	// ErrMissingPage indicates when a page is unexpectedly missing.
	ErrMissingPage = errors.New("missing page")

	// ErrZeroCreatedTime indicates when an entry's CreatedTime field is zero.
	ErrZeroCreatedTime = errors.New("created time is zero")

	// ErrDiffAuthorPubs indicates when the author public keys of an entry and its page differ.
	ErrDiffAuthorPubs = errors.New("page and entry have different author public keys")

	// ErrMissingPageKeys indicates when the PageKeys of an entry are unexpectedly missing.
	ErrMissingPageKeys = errors.New("missing page keys")

	// ErrEmptyPageKeys indicates when the PageKeys of an entry are unexpectedly zero-length.
	ErrEmptyPageKeys = errors.New("empty page keys")
)

// GetKey calculates the key from the has of the proto.Message.
func GetKey(value proto.Message) (id.ID, error) {
	valueBytes, err := proto.Marshal(value)
	if err != nil {
		return nil, err
	}
	hash := sha256.Sum256(valueBytes)
	return id.FromBytes(hash[:]), nil
}

// GetAuthorPub returns the author public key for a given document.
func GetAuthorPub(d *Document) []byte {
	switch c := d.Contents.(type) {
	case *Document_Entry:
		return c.Entry.AuthorPublicKey
	case *Document_Page:
		return c.Page.AuthorPublicKey
	case *Document_Envelope:
		return c.Envelope.AuthorPublicKey
	}
	panic(ErrUnknownDocumentType)
}

// GetEntryPageKeys returns the []id.ID page keys if the entry is multi-page. It returns nil for
// single-page entries.
func GetEntryPageKeys(entryDoc *Document) ([]id.ID, error) {
	entry, ok := entryDoc.Contents.(*Document_Entry)
	if !ok {
		return nil, ErrUnexpectedDocumentType
	}
	if entry.Entry.Page != nil {
		return nil, nil
	}
	if entry.Entry.PageKeys != nil {
		pageKeyBytes := entry.Entry.PageKeys
		pageKeys := make([]id.ID, len(pageKeyBytes))
		for i, keyBytes := range pageKeyBytes {
			pageKeys[i] = id.FromBytes(keyBytes)
		}
		return pageKeys, nil
	}
	return nil, ErrUnexpectedDocumentType
}

// GetPageDocument wraps a Page into a Document, returning it and its key.
func GetPageDocument(page *Page) (*Document, id.ID, error) {
	pageDoc := &Document{
		Contents: &Document_Page{
			Page: page,
		},
	}
	// store single page as separate doc
	pageKey, err := GetKey(pageDoc)
	cerrors.MaybePanic(err) // should never happen
	return pageDoc, pageKey, nil
}

// ValidateDocument checks that all fields of a Document are populated and have the expected
// lengths.
func ValidateDocument(d *Document) error {
	if d == nil {
		return ErrMissingDocument
	}
	switch c := d.Contents.(type) {
	case *Document_Envelope:
		return ValidateEnvelope(c.Envelope)
	case *Document_Entry:
		return ValidateEntry(c.Entry)
	case *Document_Page:
		return ValidatePage(c.Page)
	}
	return ErrUnknownDocumentType
}

// ValidateEnvelope checks that all fields of an Envelope are populated and have the expected
// lengths.
func ValidateEnvelope(e *Envelope) error {
	if e == nil {
		return ErrMissingEnvelope
	}
	if err := ValidateBytes(e.EntryKey, DocumentKeyLength, "EntryKey"); err != nil {
		return err
	}
	if err := ValidatePublicKey(e.AuthorPublicKey); err != nil {
		return err
	}
	if err := ValidatePublicKey(e.ReaderPublicKey); err != nil {
		return err
	}
	if err := ValidateBytes(e.EekCiphertext, EEKCiphertextLength, "EntryEncryptionKeys"); err != nil {
		return err
	}
	if err := ValidateHMAC256(e.EekCiphertextMac); err != nil {
		return err
	}
	return nil
}

// ValidateEntry checks that all fields of an Entry are populated and have the expected byte
// lengths.
func ValidateEntry(e *Entry) error {
	if e == nil {
		return ErrMissingEntry
	}
	if err := ValidatePublicKey(e.AuthorPublicKey); err != nil {
		return err
	}
	if e.CreatedTime == 0 {
		return ErrZeroCreatedTime
	}
	if err := ValidateHMAC256(e.MetadataCiphertextMac); err != nil {
		return err
	}
	if err := ValidateNotEmpty(e.MetadataCiphertext, "MetadataCiphertext"); err != nil {
		return err
	}
	if err := validateEntryContents(e); err != nil {
		return err
	}
	return nil
}

func validateEntryContents(e *Entry) error {
	if e.Page != nil {
		if !bytes.Equal(e.AuthorPublicKey, e.Page.AuthorPublicKey) {
			return ErrDiffAuthorPubs
		}
		return ValidatePage(e.Page)
	}
	if e.PageKeys != nil {
		return ValidatePageKeys(e.PageKeys)
	}
	return ErrUnknownDocumentType
}

// ValidatePage checks that all fields of a Page are populated and have the expected lengths.
func ValidatePage(p *Page) error {
	if p == nil {
		return ErrMissingPage
	}
	if err := ValidatePublicKey(p.AuthorPublicKey); err != nil {
		return err
	}
	// nothing to check for index, since it's zero value is legitimate
	if err := ValidateHMAC256(p.CiphertextMac); err != nil {
		return err
	}
	if err := ValidateNotEmpty(p.Ciphertext, "Ciphertext"); err != nil {
		return err
	}
	return nil
}

// ValidatePageKeys checks that all fields of a PageKeys are populated and have the expected
// lengths.
func ValidatePageKeys(pk [][]byte) error {
	if pk == nil {
		return ErrMissingPageKeys
	}
	if len(pk) == 0 {
		return ErrEmptyPageKeys
	}
	for i, k := range pk {
		if err := ValidateNotEmpty(k, fmt.Sprintf("key %d", i)); err != nil {
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

// ValidateHMACKey checks that a value can be an HMAC-256 key.
func ValidateHMACKey(value []byte) error {
	return ValidateBytes(value, HMACKeyLength, "HMACKey")
}

// ValidatePageIVSeed checks that a value can be a 256-bit initialization vector seed.
func ValidatePageIVSeed(value []byte) error {
	return ValidateBytes(value, PageIVSeedLength, "PageIVSeed")
}

// ValidateMetadataIV checks that a value can be a 12-byte GCM initialization vector.
func ValidateMetadataIV(value []byte) error {
	return ValidateBytes(value, BlockCipherIVLength, "MetadataIV")
}

// ValidateHMAC256 checks that a value can be an HMAC-256.
func ValidateHMAC256(value []byte) error {
	return ValidateBytes(value, HMAC256Length, "HMAC256")
}

// ValidateBytes returns whether the byte slice is not empty and has an expected length.
func ValidateBytes(value []byte, expectedLen int, name string) error {
	if err := ValidateNotEmpty(value, name); err != nil {
		return err
	}
	if len(value) != expectedLen {
		return fmt.Errorf("%s must have length %d, found length %d", name, expectedLen,
			len(value))
	}
	return nil
}

// ValidateNotEmpty returns whether the byte slice is not empty.
func ValidateNotEmpty(value []byte, name string) error {
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
