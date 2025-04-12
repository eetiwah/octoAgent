package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"io"
	"os"
)

func EncryptFile(password string, inputFilepath string, outputFilepath string) error {
	// Hash the password using SHA-256
	hashedPassword := sha256.Sum256([]byte(password))

	// Convert the hashed password to a byte slice
	key := hashedPassword[:]

	// Generate a random nonce
	nonce := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}

	// Open the input file
	inputFileHandle, err := os.Open(inputFilepath)
	if err != nil {
		return err
	}
	defer inputFileHandle.Close()

	// Create the output file
	outputFileHandle, err := os.Create(outputFilepath)
	if err != nil {
		return err
	}
	defer outputFileHandle.Close()

	// Write the nonce to the beginning of the output file
	if _, err := outputFileHandle.Write(nonce); err != nil {
		return err
	}

	// Create a new AES cipher block using the provided key
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Create a cipher mode for AES in CBC mode
	cbc := cipher.NewCBCEncrypter(block, nonce)

	// Create a buffer to hold the encrypted data
	buffer := make([]byte, aes.BlockSize)

	// Encrypt the data
	for {
		n, err := inputFileHandle.Read(buffer)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		// Pad the input data if necessary
		if n < aes.BlockSize {
			for i := n; i < aes.BlockSize; i++ {
				buffer[i] = byte(aes.BlockSize - n)
			}
		}

		// Encrypt the buffer
		cbc.CryptBlocks(buffer, buffer)

		// Write the encrypted data to the output file
		if _, err := outputFileHandle.Write(buffer); err != nil {
			return err
		}

		if n < aes.BlockSize {
			break
		}
	}
	return nil
}

func DecryptFile(password string, inputFilepath string, outputFilepath string) error {
	// Hash the password using SHA-256
	hashedPassword := sha256.Sum256([]byte(password))

	// Convert the hashed password to a byte slice
	key := hashedPassword[:]

	// Open the input file
	inputFileHandle, err := os.Open(inputFilepath)
	if err != nil {
		return err
	}
	defer inputFileHandle.Close()

	// Create the output file
	outputFileHandle, err := os.Create(outputFilepath)
	if err != nil {
		return err
	}
	defer outputFileHandle.Close()

	// Get the file size for reading
	stat, err := inputFileHandle.Stat()
	if err != nil {
		return err
	}
	fileSize := stat.Size()

	// Create a buffer to read the file contents
	encryptedData := make([]byte, fileSize)
	_, err = inputFileHandle.Read(encryptedData)
	if err != nil {
		return err
	}

	// Get the nonce from the beginning of the encrypted data
	nonceSize := 16 // 16 bytes for a 16-byte nonce
	nonce := encryptedData[:nonceSize]
	encryptedData = encryptedData[nonceSize:]

	// Create a new AES cipher block using the provided key
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}

	// Create a cipher mode for AES in CBC mode
	cbc := cipher.NewCBCDecrypter(block, nonce)

	// Create a buffer to hold the decrypted data
	decryptedData := make([]byte, len(encryptedData))

	// Decrypt the data
	cbc.CryptBlocks(decryptedData, encryptedData)

	// Remove padding
	paddingLen := int(decryptedData[len(decryptedData)-1])
	if paddingLen > 0 && paddingLen <= aes.BlockSize {
		decryptedData = decryptedData[:len(decryptedData)-paddingLen]
	}

	// Write the decrypted data to the output file
	_, err = outputFileHandle.Write(decryptedData)
	if err != nil {
		return err
	}

	return nil
}
