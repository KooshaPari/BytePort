package lib
type User struct{
	Name string
	Email string
	Password string
	awsCreds struct{
		accessKeyId string
		secretAccessKey string
	}
	openAICreds struct{
		apiKey string
	}
	portfolio struct{
		rootEndpoint string
		apiKey string
	}
	git struct{
		repoUrl string
		authMethod string
		authKey string
		targetDirectory string
	}
}
import (
    "crypto/rand"
    "fmt"
    "log"

    "golang.org/x/crypto/argon2"
)

type params struct {
    memory      uint32
    iterations  uint32
    parallelism uint8
    saltLength  uint32
    keyLength   uint32
}

func encryptPass() {
    

    // Establish the parameters to use for Argon2.
    p := &params{
        memory:      64 * 1024,
        iterations:  3,
        parallelism: 2,
        saltLength:  16,
        keyLength:   32,
    }

    // Pass the plaintext password and parameters to our generateFromPassword
    // helper function.
    hash, err := generateFromPassword("password123", p)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println(hash)
}

func generateFromPassword(password string, p *params) (hash []byte, err error) {
    // Generate a cryptographically secure random salt.
    salt, err := generateRandomBytes(p.saltLength)
    if err != nil {
        return nil, err
    }
    hash = argon2.IDKey([]byte(password), salt, p.iterations, p.memory, p.parallelism, p.keyLength)

    return hash, nil
}

func generateRandomBytes(n uint32) ([]byte, error) {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return nil, err
    }

    return b, nil
}