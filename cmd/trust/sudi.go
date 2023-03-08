package main

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/apex/log"
	"github.com/google/uuid"
)

// Generate sudi private key and cert.
func doSudiCert(VMname, keyset string) error {
	if VMname == "" {
		return errors.New("VM name must be provided")
	}

	// Check if the VM has been initialized
	cPath := ConfPath(VMname)
	if !PathExists(cPath) {
		return fmt.Errorf("%s has not been initialized", VMname)
	}

	// Check if a sudi key or cert already exists for the VM
	sudiDir, err := getSudiDir()
	if err != nil {
		return err
	}
	sudiPath := filepath.Join(sudiDir, VMname)
	_, err = os.Stat(filepath.Join(sudiPath, "privkey.pem"))
	if err == nil {
		fmt.Printf("A privkey.pem already exists for %s in %s.\n", VMname, sudiPath)
		return err
	}
	_, err = os.Stat(filepath.Join(sudiPath, "cert.pem"))
	if err == nil {
		fmt.Printf("A cert.pem already exists for %s in %s.\n", VMname, sudiPath)
		return err
	}

	// Prepare the cert template
	// Get this machine's UUID to add to the Subject in cert
	trustDir, err := getTrustPath()
	if err != nil {
		return err
	}
	content, err := os.ReadFile(filepath.Join(trustDir, "manifest/uuid"))
	if err != nil {
		return err
	}
	productUUID := string(content)

	certTemplate := newCertTemplate(productUUID, uuid.NewString())

	// get the CA info
	CAcert, CAprivkey, err := getCA("sudi-ca", keyset)
	if err != nil {
		return err
	}

	err = os.MkdirAll(sudiPath, 0755)
	if err != nil {
		return err
	}
	err = SignCert(&certTemplate, CAcert, CAprivkey, sudiPath)
	if err != nil {
		return err
	}
	log.Infof("Generated sudi key and cert saved in %s directory\n", sudiPath)
	return nil
}

func newCertTemplate(productUUID, machineUUID string) x509.Certificate {
	return x509.Certificate{
		Subject: pkix.Name{
			SerialNumber: fmt.Sprintf("PID:%s SN:%s", productUUID, machineUUID),
			CommonName:   machineUUID,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Date(2099, time.December, 31, 23, 0, 0, 0, time.UTC),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment | x509.KeyUsageDataEncipherment,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth},
	}
}
