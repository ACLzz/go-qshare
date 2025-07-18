// Copyright 2020 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Proto definitions for SecureMessage format

// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        v6.30.2
// source: securemessage.proto

package securemessage

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Supported "signature" schemes (both symmetric key and public key based)
type SigScheme int32

const (
	SigScheme_HMAC_SHA256       SigScheme = 1
	SigScheme_ECDSA_P256_SHA256 SigScheme = 2
	// Not recommended -- use ECDSA_P256_SHA256 instead
	SigScheme_RSA2048_SHA256 SigScheme = 3
)

// Enum value maps for SigScheme.
var (
	SigScheme_name = map[int32]string{
		1: "HMAC_SHA256",
		2: "ECDSA_P256_SHA256",
		3: "RSA2048_SHA256",
	}
	SigScheme_value = map[string]int32{
		"HMAC_SHA256":       1,
		"ECDSA_P256_SHA256": 2,
		"RSA2048_SHA256":    3,
	}
)

func (x SigScheme) Enum() *SigScheme {
	p := new(SigScheme)
	*p = x
	return p
}

func (x SigScheme) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SigScheme) Descriptor() protoreflect.EnumDescriptor {
	return file_securemessage_proto_enumTypes[0].Descriptor()
}

func (SigScheme) Type() protoreflect.EnumType {
	return &file_securemessage_proto_enumTypes[0]
}

func (x SigScheme) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *SigScheme) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = SigScheme(num)
	return nil
}

// Deprecated: Use SigScheme.Descriptor instead.
func (SigScheme) EnumDescriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{0}
}

// Supported encryption schemes
type EncScheme int32

const (
	// No encryption
	EncScheme_NONE        EncScheme = 1
	EncScheme_AES_256_CBC EncScheme = 2
)

// Enum value maps for EncScheme.
var (
	EncScheme_name = map[int32]string{
		1: "NONE",
		2: "AES_256_CBC",
	}
	EncScheme_value = map[string]int32{
		"NONE":        1,
		"AES_256_CBC": 2,
	}
)

func (x EncScheme) Enum() *EncScheme {
	p := new(EncScheme)
	*p = x
	return p
}

func (x EncScheme) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EncScheme) Descriptor() protoreflect.EnumDescriptor {
	return file_securemessage_proto_enumTypes[1].Descriptor()
}

func (EncScheme) Type() protoreflect.EnumType {
	return &file_securemessage_proto_enumTypes[1]
}

func (x EncScheme) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *EncScheme) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = EncScheme(num)
	return nil
}

// Deprecated: Use EncScheme.Descriptor instead.
func (EncScheme) EnumDescriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{1}
}

// A list of supported public key types
type PublicKeyType int32

const (
	PublicKeyType_EC_P256 PublicKeyType = 1
	PublicKeyType_RSA2048 PublicKeyType = 2
	// 2048-bit MODP group 14, from RFC 3526
	PublicKeyType_DH2048_MODP PublicKeyType = 3
)

// Enum value maps for PublicKeyType.
var (
	PublicKeyType_name = map[int32]string{
		1: "EC_P256",
		2: "RSA2048",
		3: "DH2048_MODP",
	}
	PublicKeyType_value = map[string]int32{
		"EC_P256":     1,
		"RSA2048":     2,
		"DH2048_MODP": 3,
	}
)

func (x PublicKeyType) Enum() *PublicKeyType {
	p := new(PublicKeyType)
	*p = x
	return p
}

func (x PublicKeyType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (PublicKeyType) Descriptor() protoreflect.EnumDescriptor {
	return file_securemessage_proto_enumTypes[2].Descriptor()
}

func (PublicKeyType) Type() protoreflect.EnumType {
	return &file_securemessage_proto_enumTypes[2]
}

func (x PublicKeyType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *PublicKeyType) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = PublicKeyType(num)
	return nil
}

// Deprecated: Use PublicKeyType.Descriptor instead.
func (PublicKeyType) EnumDescriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{2}
}

type SecureMessage struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Must contain a HeaderAndBody message
	HeaderAndBody []byte `protobuf:"bytes,1,req,name=header_and_body,json=headerAndBody" json:"header_and_body,omitempty"`
	// Signature of header_and_body
	Signature     []byte `protobuf:"bytes,2,req,name=signature" json:"signature,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *SecureMessage) Reset() {
	*x = SecureMessage{}
	mi := &file_securemessage_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SecureMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SecureMessage) ProtoMessage() {}

func (x *SecureMessage) ProtoReflect() protoreflect.Message {
	mi := &file_securemessage_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SecureMessage.ProtoReflect.Descriptor instead.
func (*SecureMessage) Descriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{0}
}

func (x *SecureMessage) GetHeaderAndBody() []byte {
	if x != nil {
		return x.HeaderAndBody
	}
	return nil
}

func (x *SecureMessage) GetSignature() []byte {
	if x != nil {
		return x.Signature
	}
	return nil
}

type Header struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	SignatureScheme  *SigScheme             `protobuf:"varint,1,req,name=signature_scheme,json=signatureScheme,enum=securemessage.SigScheme" json:"signature_scheme,omitempty"`
	EncryptionScheme *EncScheme             `protobuf:"varint,2,req,name=encryption_scheme,json=encryptionScheme,enum=securemessage.EncScheme" json:"encryption_scheme,omitempty"`
	// Identifies the verification key
	VerificationKeyId []byte `protobuf:"bytes,3,opt,name=verification_key_id,json=verificationKeyId" json:"verification_key_id,omitempty"`
	// Identifies the decryption key
	DecryptionKeyId []byte `protobuf:"bytes,4,opt,name=decryption_key_id,json=decryptionKeyId" json:"decryption_key_id,omitempty"`
	// Encryption may use an IV
	Iv []byte `protobuf:"bytes,5,opt,name=iv" json:"iv,omitempty"`
	// Arbitrary per-protocol public data, to be sent with the plain-text header
	PublicMetadata []byte `protobuf:"bytes,6,opt,name=public_metadata,json=publicMetadata" json:"public_metadata,omitempty"`
	// The length of some associated data this is not sent in this SecureMessage,
	// but which will be bound to the signature.
	AssociatedDataLength *uint32 `protobuf:"varint,7,opt,name=associated_data_length,json=associatedDataLength,def=0" json:"associated_data_length,omitempty"`
	unknownFields        protoimpl.UnknownFields
	sizeCache            protoimpl.SizeCache
}

// Default values for Header fields.
const (
	Default_Header_AssociatedDataLength = uint32(0)
)

func (x *Header) Reset() {
	*x = Header{}
	mi := &file_securemessage_proto_msgTypes[1]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *Header) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Header) ProtoMessage() {}

func (x *Header) ProtoReflect() protoreflect.Message {
	mi := &file_securemessage_proto_msgTypes[1]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Header.ProtoReflect.Descriptor instead.
func (*Header) Descriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{1}
}

func (x *Header) GetSignatureScheme() SigScheme {
	if x != nil && x.SignatureScheme != nil {
		return *x.SignatureScheme
	}
	return SigScheme_HMAC_SHA256
}

func (x *Header) GetEncryptionScheme() EncScheme {
	if x != nil && x.EncryptionScheme != nil {
		return *x.EncryptionScheme
	}
	return EncScheme_NONE
}

func (x *Header) GetVerificationKeyId() []byte {
	if x != nil {
		return x.VerificationKeyId
	}
	return nil
}

func (x *Header) GetDecryptionKeyId() []byte {
	if x != nil {
		return x.DecryptionKeyId
	}
	return nil
}

func (x *Header) GetIv() []byte {
	if x != nil {
		return x.Iv
	}
	return nil
}

func (x *Header) GetPublicMetadata() []byte {
	if x != nil {
		return x.PublicMetadata
	}
	return nil
}

func (x *Header) GetAssociatedDataLength() uint32 {
	if x != nil && x.AssociatedDataLength != nil {
		return *x.AssociatedDataLength
	}
	return Default_Header_AssociatedDataLength
}

type HeaderAndBody struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Public data about this message (to be bound in the signature)
	Header *Header `protobuf:"bytes,1,req,name=header" json:"header,omitempty"`
	// Payload data
	Body          []byte `protobuf:"bytes,2,req,name=body" json:"body,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HeaderAndBody) Reset() {
	*x = HeaderAndBody{}
	mi := &file_securemessage_proto_msgTypes[2]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeaderAndBody) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeaderAndBody) ProtoMessage() {}

func (x *HeaderAndBody) ProtoReflect() protoreflect.Message {
	mi := &file_securemessage_proto_msgTypes[2]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeaderAndBody.ProtoReflect.Descriptor instead.
func (*HeaderAndBody) Descriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{2}
}

func (x *HeaderAndBody) GetHeader() *Header {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *HeaderAndBody) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

// Must be kept wire-format compatible with HeaderAndBody. Provides the
// SecureMessage code with a consistent wire-format representation that
// remains stable irrespective of protobuf implementation choices. This
// low-level representation of a HeaderAndBody should not be used by
// any code outside of the SecureMessage library implementation/tests.
type HeaderAndBodyInternal struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// A raw (wire-format) byte encoding of a Header, suitable for hashing
	Header []byte `protobuf:"bytes,1,req,name=header" json:"header,omitempty"`
	// Payload data
	Body          []byte `protobuf:"bytes,2,req,name=body" json:"body,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *HeaderAndBodyInternal) Reset() {
	*x = HeaderAndBodyInternal{}
	mi := &file_securemessage_proto_msgTypes[3]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *HeaderAndBodyInternal) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*HeaderAndBodyInternal) ProtoMessage() {}

func (x *HeaderAndBodyInternal) ProtoReflect() protoreflect.Message {
	mi := &file_securemessage_proto_msgTypes[3]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use HeaderAndBodyInternal.ProtoReflect.Descriptor instead.
func (*HeaderAndBodyInternal) Descriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{3}
}

func (x *HeaderAndBodyInternal) GetHeader() []byte {
	if x != nil {
		return x.Header
	}
	return nil
}

func (x *HeaderAndBodyInternal) GetBody() []byte {
	if x != nil {
		return x.Body
	}
	return nil
}

// A convenience proto for encoding NIST P-256 elliptic curve public keys
type EcP256PublicKey struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// x and y are encoded in big-endian two's complement (slightly wasteful)
	// Client MUST verify (x,y) is a valid point on NIST P256
	X             []byte `protobuf:"bytes,1,req,name=x" json:"x,omitempty"`
	Y             []byte `protobuf:"bytes,2,req,name=y" json:"y,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *EcP256PublicKey) Reset() {
	*x = EcP256PublicKey{}
	mi := &file_securemessage_proto_msgTypes[4]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *EcP256PublicKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*EcP256PublicKey) ProtoMessage() {}

func (x *EcP256PublicKey) ProtoReflect() protoreflect.Message {
	mi := &file_securemessage_proto_msgTypes[4]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use EcP256PublicKey.ProtoReflect.Descriptor instead.
func (*EcP256PublicKey) Descriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{4}
}

func (x *EcP256PublicKey) GetX() []byte {
	if x != nil {
		return x.X
	}
	return nil
}

func (x *EcP256PublicKey) GetY() []byte {
	if x != nil {
		return x.Y
	}
	return nil
}

// A convenience proto for encoding RSA public keys with small exponents
type SimpleRsaPublicKey struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Encoded in big-endian two's complement
	N             []byte `protobuf:"bytes,1,req,name=n" json:"n,omitempty"`
	E             *int32 `protobuf:"varint,2,opt,name=e,def=65537" json:"e,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

// Default values for SimpleRsaPublicKey fields.
const (
	Default_SimpleRsaPublicKey_E = int32(65537)
)

func (x *SimpleRsaPublicKey) Reset() {
	*x = SimpleRsaPublicKey{}
	mi := &file_securemessage_proto_msgTypes[5]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *SimpleRsaPublicKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SimpleRsaPublicKey) ProtoMessage() {}

func (x *SimpleRsaPublicKey) ProtoReflect() protoreflect.Message {
	mi := &file_securemessage_proto_msgTypes[5]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SimpleRsaPublicKey.ProtoReflect.Descriptor instead.
func (*SimpleRsaPublicKey) Descriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{5}
}

func (x *SimpleRsaPublicKey) GetN() []byte {
	if x != nil {
		return x.N
	}
	return nil
}

func (x *SimpleRsaPublicKey) GetE() int32 {
	if x != nil && x.E != nil {
		return *x.E
	}
	return Default_SimpleRsaPublicKey_E
}

// A convenience proto for encoding Diffie-Hellman public keys,
// for use only when Elliptic Curve based key exchanges are not possible.
// (Note that the group parameters must be specified separately)
type DhPublicKey struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// Big-endian two's complement encoded group element
	Y             []byte `protobuf:"bytes,1,req,name=y" json:"y,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *DhPublicKey) Reset() {
	*x = DhPublicKey{}
	mi := &file_securemessage_proto_msgTypes[6]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *DhPublicKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DhPublicKey) ProtoMessage() {}

func (x *DhPublicKey) ProtoReflect() protoreflect.Message {
	mi := &file_securemessage_proto_msgTypes[6]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DhPublicKey.ProtoReflect.Descriptor instead.
func (*DhPublicKey) Descriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{6}
}

func (x *DhPublicKey) GetY() []byte {
	if x != nil {
		return x.Y
	}
	return nil
}

type GenericPublicKey struct {
	state            protoimpl.MessageState `protogen:"open.v1"`
	Type             *PublicKeyType         `protobuf:"varint,1,req,name=type,enum=securemessage.PublicKeyType" json:"type,omitempty"`
	EcP256PublicKey  *EcP256PublicKey       `protobuf:"bytes,2,opt,name=ec_p256_public_key,json=ecP256PublicKey" json:"ec_p256_public_key,omitempty"`
	Rsa2048PublicKey *SimpleRsaPublicKey    `protobuf:"bytes,3,opt,name=rsa2048_public_key,json=rsa2048PublicKey" json:"rsa2048_public_key,omitempty"`
	// Use only as a last resort
	Dh2048PublicKey *DhPublicKey `protobuf:"bytes,4,opt,name=dh2048_public_key,json=dh2048PublicKey" json:"dh2048_public_key,omitempty"`
	unknownFields   protoimpl.UnknownFields
	sizeCache       protoimpl.SizeCache
}

func (x *GenericPublicKey) Reset() {
	*x = GenericPublicKey{}
	mi := &file_securemessage_proto_msgTypes[7]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *GenericPublicKey) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*GenericPublicKey) ProtoMessage() {}

func (x *GenericPublicKey) ProtoReflect() protoreflect.Message {
	mi := &file_securemessage_proto_msgTypes[7]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use GenericPublicKey.ProtoReflect.Descriptor instead.
func (*GenericPublicKey) Descriptor() ([]byte, []int) {
	return file_securemessage_proto_rawDescGZIP(), []int{7}
}

func (x *GenericPublicKey) GetType() PublicKeyType {
	if x != nil && x.Type != nil {
		return *x.Type
	}
	return PublicKeyType_EC_P256
}

func (x *GenericPublicKey) GetEcP256PublicKey() *EcP256PublicKey {
	if x != nil {
		return x.EcP256PublicKey
	}
	return nil
}

func (x *GenericPublicKey) GetRsa2048PublicKey() *SimpleRsaPublicKey {
	if x != nil {
		return x.Rsa2048PublicKey
	}
	return nil
}

func (x *GenericPublicKey) GetDh2048PublicKey() *DhPublicKey {
	if x != nil {
		return x.Dh2048PublicKey
	}
	return nil
}

var File_securemessage_proto protoreflect.FileDescriptor

const file_securemessage_proto_rawDesc = "" +
	"\n" +
	"\x13securemessage.proto\x12\rsecuremessage\"U\n" +
	"\rSecureMessage\x12&\n" +
	"\x0fheader_and_body\x18\x01 \x02(\fR\rheaderAndBody\x12\x1c\n" +
	"\tsignature\x18\x02 \x02(\fR\tsignature\"\xe2\x02\n" +
	"\x06Header\x12C\n" +
	"\x10signature_scheme\x18\x01 \x02(\x0e2\x18.securemessage.SigSchemeR\x0fsignatureScheme\x12E\n" +
	"\x11encryption_scheme\x18\x02 \x02(\x0e2\x18.securemessage.EncSchemeR\x10encryptionScheme\x12.\n" +
	"\x13verification_key_id\x18\x03 \x01(\fR\x11verificationKeyId\x12*\n" +
	"\x11decryption_key_id\x18\x04 \x01(\fR\x0fdecryptionKeyId\x12\x0e\n" +
	"\x02iv\x18\x05 \x01(\fR\x02iv\x12'\n" +
	"\x0fpublic_metadata\x18\x06 \x01(\fR\x0epublicMetadata\x127\n" +
	"\x16associated_data_length\x18\a \x01(\r:\x010R\x14associatedDataLength\"R\n" +
	"\rHeaderAndBody\x12-\n" +
	"\x06header\x18\x01 \x02(\v2\x15.securemessage.HeaderR\x06header\x12\x12\n" +
	"\x04body\x18\x02 \x02(\fR\x04body\"C\n" +
	"\x15HeaderAndBodyInternal\x12\x16\n" +
	"\x06header\x18\x01 \x02(\fR\x06header\x12\x12\n" +
	"\x04body\x18\x02 \x02(\fR\x04body\"-\n" +
	"\x0fEcP256PublicKey\x12\f\n" +
	"\x01x\x18\x01 \x02(\fR\x01x\x12\f\n" +
	"\x01y\x18\x02 \x02(\fR\x01y\"7\n" +
	"\x12SimpleRsaPublicKey\x12\f\n" +
	"\x01n\x18\x01 \x02(\fR\x01n\x12\x13\n" +
	"\x01e\x18\x02 \x01(\x05:\x0565537R\x01e\"\x1b\n" +
	"\vDhPublicKey\x12\f\n" +
	"\x01y\x18\x01 \x02(\fR\x01y\"\xaa\x02\n" +
	"\x10GenericPublicKey\x120\n" +
	"\x04type\x18\x01 \x02(\x0e2\x1c.securemessage.PublicKeyTypeR\x04type\x12K\n" +
	"\x12ec_p256_public_key\x18\x02 \x01(\v2\x1e.securemessage.EcP256PublicKeyR\x0fecP256PublicKey\x12O\n" +
	"\x12rsa2048_public_key\x18\x03 \x01(\v2!.securemessage.SimpleRsaPublicKeyR\x10rsa2048PublicKey\x12F\n" +
	"\x11dh2048_public_key\x18\x04 \x01(\v2\x1a.securemessage.DhPublicKeyR\x0fdh2048PublicKey*G\n" +
	"\tSigScheme\x12\x0f\n" +
	"\vHMAC_SHA256\x10\x01\x12\x15\n" +
	"\x11ECDSA_P256_SHA256\x10\x02\x12\x12\n" +
	"\x0eRSA2048_SHA256\x10\x03*&\n" +
	"\tEncScheme\x12\b\n" +
	"\x04NONE\x10\x01\x12\x0f\n" +
	"\vAES_256_CBC\x10\x02*:\n" +
	"\rPublicKeyType\x12\v\n" +
	"\aEC_P256\x10\x01\x12\v\n" +
	"\aRSA2048\x10\x02\x12\x0f\n" +
	"\vDH2048_MODP\x10\x03B\x8b\x01\n" +
	"/com.google.security.cryptauth.lib.securemessageB\x12SecureMessageProtoH\x03Z;github.com/ACLzz/qshare/internal/protobuf/gen/securemessage\xa2\x02\x04SMSG"

var (
	file_securemessage_proto_rawDescOnce sync.Once
	file_securemessage_proto_rawDescData []byte
)

func file_securemessage_proto_rawDescGZIP() []byte {
	file_securemessage_proto_rawDescOnce.Do(func() {
		file_securemessage_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_securemessage_proto_rawDesc), len(file_securemessage_proto_rawDesc)))
	})
	return file_securemessage_proto_rawDescData
}

var file_securemessage_proto_enumTypes = make([]protoimpl.EnumInfo, 3)
var file_securemessage_proto_msgTypes = make([]protoimpl.MessageInfo, 8)
var file_securemessage_proto_goTypes = []any{
	(SigScheme)(0),                // 0: securemessage.SigScheme
	(EncScheme)(0),                // 1: securemessage.EncScheme
	(PublicKeyType)(0),            // 2: securemessage.PublicKeyType
	(*SecureMessage)(nil),         // 3: securemessage.SecureMessage
	(*Header)(nil),                // 4: securemessage.Header
	(*HeaderAndBody)(nil),         // 5: securemessage.HeaderAndBody
	(*HeaderAndBodyInternal)(nil), // 6: securemessage.HeaderAndBodyInternal
	(*EcP256PublicKey)(nil),       // 7: securemessage.EcP256PublicKey
	(*SimpleRsaPublicKey)(nil),    // 8: securemessage.SimpleRsaPublicKey
	(*DhPublicKey)(nil),           // 9: securemessage.DhPublicKey
	(*GenericPublicKey)(nil),      // 10: securemessage.GenericPublicKey
}
var file_securemessage_proto_depIdxs = []int32{
	0, // 0: securemessage.Header.signature_scheme:type_name -> securemessage.SigScheme
	1, // 1: securemessage.Header.encryption_scheme:type_name -> securemessage.EncScheme
	4, // 2: securemessage.HeaderAndBody.header:type_name -> securemessage.Header
	2, // 3: securemessage.GenericPublicKey.type:type_name -> securemessage.PublicKeyType
	7, // 4: securemessage.GenericPublicKey.ec_p256_public_key:type_name -> securemessage.EcP256PublicKey
	8, // 5: securemessage.GenericPublicKey.rsa2048_public_key:type_name -> securemessage.SimpleRsaPublicKey
	9, // 6: securemessage.GenericPublicKey.dh2048_public_key:type_name -> securemessage.DhPublicKey
	7, // [7:7] is the sub-list for method output_type
	7, // [7:7] is the sub-list for method input_type
	7, // [7:7] is the sub-list for extension type_name
	7, // [7:7] is the sub-list for extension extendee
	0, // [0:7] is the sub-list for field type_name
}

func init() { file_securemessage_proto_init() }
func file_securemessage_proto_init() {
	if File_securemessage_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_securemessage_proto_rawDesc), len(file_securemessage_proto_rawDesc)),
			NumEnums:      3,
			NumMessages:   8,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_securemessage_proto_goTypes,
		DependencyIndexes: file_securemessage_proto_depIdxs,
		EnumInfos:         file_securemessage_proto_enumTypes,
		MessageInfos:      file_securemessage_proto_msgTypes,
	}.Build()
	File_securemessage_proto = out.File
	file_securemessage_proto_goTypes = nil
	file_securemessage_proto_depIdxs = nil
}
