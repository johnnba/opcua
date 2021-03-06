// Copyright 2018-2019 opcua authors. All rights reserved.
// Use of this source code is governed by a MIT-style license that can be
// found in the LICENSE file.

// Package uapolicy implements the encryption, decryption, signing,
// and signature verifying algorithms for Security Policy profiles as
// defined in Part 7 of the OPC-UA specifications (version 1.04)
package uapolicy

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"sort"
)

const (
	SecurityPolicyURL                 = "http://opcfoundation.org/UA/SecurityPolicy#"
	SecurityPolicyNone                = "http://opcfoundation.org/UA/SecurityPolicy#None"
	SecurityPolicyBasic128Rsa15       = "http://opcfoundation.org/UA/SecurityPolicy#Basic128Rsa15"
	SecurityPolicyBasic256            = "http://opcfoundation.org/UA/SecurityPolicy#Basic256"
	SecurityPolicyBasic256Sha256      = "http://opcfoundation.org/UA/SecurityPolicy#Basic256Sha256"
	SecurityPolicyAes128Sha256RsaOaep = "http://opcfoundation.org/UA/SecurityPolicy#Aes128_Sha256_RsaOaep"
	SecurityPolicyAes256Sha256RsaPss  = "http://opcfoundation.org/UA/SecurityPolicy#Aes256_Sha256_RsaPss"
)

// SupportedPolicies returns all supported Security Policies
// (and therefore, valid inputs to Asymmetric(...) and Symmetric(...))
func SupportedPolicies() []string {
	var uris []string
	for k := range policies {
		uris = append(uris, k)
	}
	sort.Strings(uris)
	return uris
}

// Asymmetric returns the asymmetric encryption algorithm for the given security policy.
func Asymmetric(uri string, localKey *rsa.PrivateKey, remoteKey *rsa.PublicKey) (*EncryptionAlgorithm, error) {
	p, ok := policies[uri]
	if !ok {
		return nil, fmt.Errorf("unsupported security policy %s", uri)
	}

	if uri != SecurityPolicyNone && (localKey == nil || remoteKey == nil) {
		return nil, errors.New("invalid asymmetric security policy config: both keys required")
	}

	return p.asymmetric(localKey, remoteKey)
}

// Symmetrics returns the symmetric encryption algorithm for the given security policy.
func Symmetric(uri string, localNonce, remoteNonce []byte) (*EncryptionAlgorithm, error) {
	p, ok := policies[uri]
	if !ok {
		return nil, fmt.Errorf("unsupported security policy %s", uri)
	}

	if uri != SecurityPolicyNone && (localNonce == nil || remoteNonce == nil) {
		return nil, errors.New("invalid symmetric security policy config: both nonces required")
	}

	return p.symmetric(localNonce, remoteNonce)
}

// EncryptionAlgorithm wraps the functions used to return the various
// methods required to implement the symmetric and asymmetric algorithms
// Function variables were used instead of an interface to make better use
// of policies which implement the same algorithms in different combinations
//
// EncryptionAlgorithm should always be instantiated through calls to
// SecurityPolicy.Symmetric() and SecurityPolicy.Asymmetric() to ensure
// correct behavior.
//
// The zero value of this struct will use SecurityPolicy#None although
// using in this manner is discouraged for readability
type EncryptionAlgorithm struct {
	blockSize           int
	plainttextBlockSize int
	decrypt             interface{ Decrypt([]byte) ([]byte, error) }
	encrypt             interface{ Encrypt([]byte) ([]byte, error) }
	signature           interface{ Signature([]byte) ([]byte, error) }
	verifySignature     interface{ Verify([]byte, []byte) error }
	nonceLength         int
	signatureLength     int
	encryptionURI       string
	signatureURI        string
}

// BlockSize returns the underlying encryption algorithm's blocksize.
// Used to calculate the padding required to make the cleartext an
// even multiple of the blocksize
func (e *EncryptionAlgorithm) BlockSize() int {
	return e.blockSize
}

// PlaintextBlockSize returns the size of the plaintext blocksize that
// can be fed into the encryption algorithm.
// Used to calculate the amount of padding to add to the
// unencrypted message
func (e *EncryptionAlgorithm) PlaintextBlockSize() int {
	return e.plainttextBlockSize
}

// Encrypt encrypts the input cleartext based on the algorithms and keys passed in
func (e *EncryptionAlgorithm) Encrypt(cleartext []byte) (ciphertext []byte, err error) {
	if e.encrypt == nil {
		e.encrypt = &None{}
	}

	return e.encrypt.Encrypt(cleartext)
}

// Decrypt decrypts the input ciphertext based on the algorithms and keys passed in
func (e *EncryptionAlgorithm) Decrypt(ciphertext []byte) (cleartext []byte, err error) {
	if e.decrypt == nil {
		e.decrypt = &None{}
	}

	return e.decrypt.Decrypt(ciphertext)
}

// Signature returns the cryptographic signature of message
func (e *EncryptionAlgorithm) Signature(message []byte) (signature []byte, err error) {
	if e.signature == nil {
		e.signature = &None{}
	}

	return e.signature.Signature(message)
}

// VerifySignature validates that 'signature' is the correct cryptographic signature
// of 'message' or returns an error.
// A return value of nil means the signature is valid
func (e *EncryptionAlgorithm) VerifySignature(message, signature []byte) error {
	if e.verifySignature == nil {
		e.verifySignature = &None{}
	}

	return e.verifySignature.Verify(message, signature)
}

// SignatureLength returns the length in bytes for the signature algorithm
func (e *EncryptionAlgorithm) SignatureLength() int {
	return e.signatureLength
}

// NonceLength returns the recommended nonce length in bytes for the security policy
// Only applicable for the Asymmetric security algorithm.  Symmetric algorithms should
// report NonceLength as zero
func (e *EncryptionAlgorithm) NonceLength() int {
	return e.nonceLength
}

// EncryptionURI returns the URI for the encryption algorithm as defined
// by the OPC-UA profiles in Part 7
func (e *EncryptionAlgorithm) EncryptionURI() string {
	return e.encryptionURI
}

// SignatureURI returns the URI for the signature algorithm as defined
// by the OPC-UA profiles in Part 7
func (e *EncryptionAlgorithm) SignatureURI() string {
	return e.signatureURI
}

var policies = map[string]policy{
	SecurityPolicyNone:                {newNoneAsymmetric, newNoneSymmetric},
	SecurityPolicyBasic128Rsa15:       {newBasic128Rsa15Asymmetric, newBasic128Rsa15Symmetric},
	SecurityPolicyBasic256:            {newBasic256Asymmetric, newBasic256Symmetric},
	SecurityPolicyBasic256Sha256:      {newBasic256Rsa256Asymmetric, newBasic256Rsa256Symmetric},
	SecurityPolicyAes128Sha256RsaOaep: {newAes128Sha256RsaOaepAsymmetric, newAes128Sha256RsaOaepSymmetric},
	SecurityPolicyAes256Sha256RsaPss:  {newAes256Sha256RsaPssAsymmetric, newAes256Sha256RsaPssSymmetric},
}

type policy struct {
	asymmetric func(localKey *rsa.PrivateKey, remoteKey *rsa.PublicKey) (*EncryptionAlgorithm, error)
	symmetric  func(localNonce []byte, remoteNonce []byte) (*EncryptionAlgorithm, error)
}
