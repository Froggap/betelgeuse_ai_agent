package authadapters

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	authports "scrapper-ai/modules/auth/applicastion/ports"
	globalerrors "scrapper-ai/modules/shared/infrastructure/errors"
	"strings"

	"golang.org/x/crypto/argon2"
)

type EncryptorAdapter struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

func NewEncryptorAdapter() authports.EncryptorPort {
	return &EncryptorAdapter{
		memory:      64 * 1024,
		iterations:  3,
		parallelism: 2,
		saltLength:  16,
		keyLength:   32,
	}
}
func (e *EncryptorAdapter) Encode(password string) (string, error) {
	salt := make([]byte, e.saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", globalerrors.NewAppError(
			500,
			"Encode Error",
			"error generando salt aleatorio",
			err,
		)
	}
	hash := argon2.IDKey(
		[]byte(password),
		salt,
		e.iterations,
		e.memory,
		e.parallelism,
		e.keyLength,
	)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	encodedHash := fmt.Sprintf(
		"$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, e.memory, e.iterations, e.parallelism, b64Salt, b64Hash,
	)

	return encodedHash, nil
}
func (e *EncryptorAdapter) Compare(password string, hashedPassword string) (bool, error) {

	parts := strings.Split(hashedPassword, "$")
	if len(parts) != 6 {
		return false, globalerrors.NewAppError(
			500,
			"Compare Error",
			"formato de hash invalido",
			nil,
		)
	}

	if parts[1] != "argon2id" {
		return false, globalerrors.NewAppError(
			500,
			"Compare Error",
			"algoritmo de encriptacion no compatible",
			nil,
		)
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil {
		return false, globalerrors.NewAppError(
			500,
			"Compare Error",
			"error parseando version del hash",
			err,
		)
	}

	var memory, iterations uint32
	var parallelism uint8
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &parallelism)
	if err != nil {
		return false, globalerrors.NewAppError(
			500,
			"Compare Error",
			"error parseando parametros del hash (memory, iterations, parallelism)",
			err,
		)
	}

	// 2. Decodificar el salt y el hash original desde Base64
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, globalerrors.NewAppError(
			500,
			"Compare Error",
			"error decodificando salt del hash",
			err,
		)
	}

	decodedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, globalerrors.NewAppError(
			500,
			"Compare Error",
			"error decodificando hash",
			err,
		)
	}

	// 3. Generar hash de prueba usando los parámetros y el salt extraídos del string
	comparisonHash := argon2.IDKey(
		[]byte(password),
		salt,
		iterations,
		memory,
		parallelism,
		uint32(len(decodedHash)),
	)

	// 4. Comparación en tiempo constante para mitigar timing attacks
	if subtle.ConstantTimeCompare(decodedHash, comparisonHash) == 1 {
		return true, nil
	}

	return false, nil
}
