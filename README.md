![Umbra](./assets/umbra-logo.jpg)

# Umbra

**Umbra** is a command-line tool that securely splits, encrypts, and redundantly stores files across multiple anonymous paste services using a zero-knowledge architecture.

## Overview

Umbra reimagines file storage by fragmenting files into encrypted chunks and distributing redundant copies across multiple anonymous paste services. Instead of traditional cloud storage, Umbra uses paste bins — services like termbin, clbin, and paste.c-net.org that are typically used for sharing code snippets or logs. These services require no authentication and leave no trace of who uploaded the data.

In Umbra's architecture, these paste services are called **providers**. This approach provides privacy, redundancy, and vendor independence — no single provider has complete information about your data.

### Key Features

- **File Fragmentation**: Split files into configurable chunks for distributed storage
- **Strong Encryption**: XChaCha20-Poly1305 authenticated encryption with Argon2id key derivation
- **Redundant Storage**: Configure multiple copies per chunk across different providers
- **Zero-Knowledge Manifest**: Encrypted metadata reveals nothing without the password
- **Ghost Modes**: Hide manifest data within images or QR codes for covert storage
- **Manifest Upload**: Upload manifests directly to providers for fully remote storage
- **Anonymous Providers**: Uses paste services that require no authentication or user tracking
- **Integrity Verification**: SHA-256 hashing ensures data integrity at chunk and file level
- **Fail-Safe Design**: Abort on any integrity mismatch to prevent data corruption

## Installation

### From Source

```bash
git clone https://github.com/henomis/umbra.git
cd umbra
make build
```

The binary will be created in `bin/umbra`.

### Install to GOPATH

```bash
make install
```

## Usage

### List Available Providers

Display all available storage providers (anonymous paste services):

```bash
umbra providers
```

This shows the list of built-in providers that can be used for chunk storage or manifest upload. Each provider is an anonymous paste service that stores your encrypted chunks.

### Upload a File

Split and encrypt a file, then distribute chunks across multiple providers:

```bash
umbra upload \
  --file ./secret.tar.gz \
  --password "your-secure-password" \
  --manifest ./secret.umbra \
  --chunks 3 \
  --copies 2 \
  --providers termbin,clbin
```

The `--providers` flag specifies which paste services to use for storing chunks. You can list multiple providers separated by commas.

**Upload manifest to a provider** (instead of saving locally):

```bash
umbra upload \
  --file ./secret.tar.gz \
  --password "your-secure-password" \
  --manifest "provider:termbin" \
  --chunks 3 \
  --copies 2
```

When using `provider:<name>`, the manifest is uploaded to the specified provider and the URL is displayed.

**Options:**

- `--file, -f`: File to upload (required)
- `--password, -p`: Encryption password (required)
- `--manifest, -m`: Path to save manifest file, or `provider:<name>` to upload to provider (required)
- `--chunk-size, -s`: Chunk size in bytes (mutually exclusive with --chunks)
- `--chunks, -c`: Number of chunks to create (default: 3, mutually exclusive with --chunk-size)
- `--copies, -n`: Number of redundant copies per chunk (default: 1)
- `--providers, -P`: Comma-separated list of providers (defaults to all available)
- `--ghost, -g`: Embed manifest in ghost mode - `image` or `qrcode` (optional)
- `--quiet, -q`: Suppress progress output

### Download a File

Reconstruct a file from its manifest:

```bash
umbra download \
  --manifest ./secret.umbra \
  --password "your-secure-password" \
  --file ./secret-restored.tar.gz
```

**Download from a provider** (if manifest was uploaded to a provider):

```bash
umbra download \
  --manifest "provider:termbin:aHR0cHM6Ly90ZXJtYmluLmNvbS94eHh4" \
  --password "your-secure-password" \
  --file ./secret-restored.tar.gz
```

When using `provider:<provider>:<hash>` format, the manifest is downloaded from the specified provider using the hash (base64-encoded metadata).

**Options:**

- `--manifest, -m`: Path to the manifest file, or `provider:<provider>:<hash>` to download from provider (required)
- `--password, -p`: Decryption password (required)
- `--file, -f`: Output file path (required)
- `--ghost, -g`: Decode manifest from ghost mode - `image` or `qrcode` (optional)
- `--quiet, -q`: Suppress progress output

### Display Manifest Information

View metadata about an encrypted manifest:

```bash
umbra info \
  --manifest ./secret.umbra \
  --password "your-secure-password"
```

**Display info for a manifest stored on a provider:**

```bash
umbra info \
  --manifest "provider:termbin:aHR0cHM6Ly90ZXJtYmluLmNvbS94eHh4" \
  --password "your-secure-password"
```

**Options:**

- `--manifest, -m`: Path to the manifest file, or `provider:<provider>:<hash>` to download from provider (required)
- `--password, -p`: Password to decrypt manifest (required)

### List Providers

View all available storage providers:

```bash
umbra providers
```

**Output Example:**
```
Available providers:
  - termbin
  - clbin
  - pipfi
  - pastecnetorg
```

Each listed provider is an anonymous paste service. These can be used with:
- The `--providers` flag to specify which services store your encrypted chunks
- Manifest upload syntax `provider:<name>` to store the manifest itself on a provider

## How It Works

### 1. File Fragmentation

Umbra divides your file into chunks using either:
- **Explicit chunk size**: Specify exact bytes per chunk (e.g., `--chunk-size 1048576` for 1MB chunks)
- **Chunk count**: Let Umbra calculate size based on number of chunks (e.g., `--chunks 5`)

Each chunk is independently hashed using SHA-256 for integrity verification.

### 2. Encryption

Before any data leaves your machine:

- **Key Derivation**: Password → encryption key via Argon2id (4 iterations, 64 MiB memory, parallelism of 4)
- **Authenticated Encryption**: XChaCha20-Poly1305 provides confidentiality and authenticity
- **Random Nonces**: Each manifest uses unique salt and nonce values

Providers (paste services) only receive encrypted blobs — they never see your plaintext data or know what file the chunks belong to.

### 3. Redundant Distribution

Chunks are uploaded to multiple providers based on your `--copies` setting:

- **Resilience**: File survives provider downtime or data loss
- **No Vendor Lock-in**: Distribute across different anonymous paste services
- **Failure Recovery**: Download succeeds if any redundant copy is available

### 4. Zero-Knowledge Manifest

The manifest file contains all reconstruction information:

- **Public Header**: Magic bytes, version, cryptographic parameters (KDF, cipher, salt, nonce)
- **Encrypted Payload**: Chunk hashes, provider identifiers, provider metadata

Without the password, the manifest reveals **nothing** about file contents, structure, or storage locations.

**Manifest Storage Options:**

- **Local File**: Save to local filesystem (e.g., `--manifest ./secret.umbra`)
- **Provider Upload**: Upload directly to a provider (e.g., `--manifest "provider:termbin"`)
- **Provider Download**: Download from a provider (e.g., `--manifest "provider:termbin:<hash>"`)
- **Ghost Mode**: Embed in an image or QR code for steganographic storage (see Ghost Modes section)

### 5. Ghost Modes (Steganography)

For covert storage, Umbra can hide the manifest inside innocent-looking images:

- **Image Mode**: Embeds manifest data into a randomly generated noise image using LSB (Least Significant Bit) steganography. The manifest is hidden in the RGB channels of the pixels.
- **QR Code Mode**: Encodes the manifest as a QR code image (max ~2.9 KB). The data is base64-encoded before embedding.

**Usage:**

```bash
# Upload with image-based manifest
umbra upload -f secret.tar.gz -p "password" -m manifest.png --ghost image

# Upload with QR code manifest
umbra upload -f secret.tar.gz -p "password" -m manifest.png --ghost qrcode

# Download from ghost manifest
umbra download -m manifest.png -p "password" -f restored.tar.gz --ghost image
```

Ghost modes provide plausible deniability — the manifest appears as an ordinary image or QR code.

### 6. Reconstruction

Download process:

1. Extract and decrypt manifest (from file or ghost image) using password
2. Fetch chunks from providers (trying redundant copies on failure)
3. Verify each chunk hash
4. Reassemble chunks in order
5. Verify final file hash

Any integrity mismatch causes immediate abort — corrupted data is never delivered.

## Storage Providers

Umbra uses anonymous paste services as storage providers. These are services designed for sharing text snippets that require no authentication and provide anonymous usage.

### Built-in Providers

- **termbin**: TCP-based paste service (termbin.com)
- **clbin**: HTTP paste service (clbin.com)
- **pipfi**: HTTP paste service (p.ip.fi)
- **pastecnetorg**: HTTP paste service (paste.c-net.org)

### Provider Constraints

Each provider has inherent limitations:
- **Maximum chunk size**: Typically 10MB per chunk
- **Expiration**: Content may expire after 30 days or less (varies by service)
- **No authentication**: Services are completely anonymous (no accounts, no tracking)
- **No guarantees**: Free public services with no SLAs or persistence guarantees
- **Terms of Service**: Each provider has its own terms of use — review them independently

Umbra validates chunk sizes against provider limits before upload and uses redundancy to mitigate expiration risks.

**Important**: Umbra is not affiliated with, endorsed by, or associated with any of the paste service providers. Users are responsible for complying with each provider's terms of service.

### Default Providers

If the `--providers` flag is omitted, Umbra distributes chunks across all available providers:

```
termbin,clbin,pipfi,pastecnetorg
```

You can specify a subset with: `--providers termbin,clbin`

## Architecture

### Design Principles

- **Security by Design**: Encryption before data transmission
- **Fail Loudly**: Abort on integrity failures rather than deliver corrupted data
- **Modularity**: Clean separation between crypto, content, and provider layers
- **Extensibility**: Simple provider interface for adding new anonymous paste services
- **Type Safety**: Strong typing catches configuration errors early

The provider abstraction allows Umbra to treat different paste services uniformly.

## Security Considerations

### Cryptographic Guarantees

- **Argon2id KDF**: Memory-hard password hashing resistant to GPU/ASIC attacks
- **XChaCha20-Poly1305 AEAD**: Modern authenticated encryption (confidentiality + integrity)
- **Random Nonces**: Each manifest uses cryptographically random salt and nonce
- **Hash Verification**: SHA-256 checksums prevent undetected corruption

### Threat Model

**Protected Against:**
- **Passive provider surveillance**: Data encrypted at rest on paste services
- **Provider data breaches**: Providers store only encrypted ciphertext
- **Data corruption**: Cryptographic verification at every step
- **Single provider failure**: Redundancy across multiple services
- **Provider curiosity**: No single provider has enough information to reconstruct your file

**NOT Protected Against:**
- **Weak passwords**: Use strong, unique passwords (brute-force is still possible)
- **Compromised endpoints**: Malware on your machine can capture plaintext
- **Manifest + password exposure**: Keep password separate from manifest
- **Traffic analysis**: Timing/size metadata visible to providers and network observers
- **Data expiration**: Providers may delete content after days/weeks without warning

### Best Practices

1. **Strong Passwords**: Use long, random passwords (consider a password manager)
2. **Separate Storage**: Store manifest and password in different locations
3. **Additional Backups**: Umbra is experimental — maintain traditional backups
4. **Secure Deletion**: Securely delete source files after upload if needed
5. **Verify Downloads**: Always check that downloaded files match expectations


## License

See [LICENSE](LICENSE) file for details.

## Disclaimer

⚠️ **Umbra is experimental software.**

- **Not a backup solution**: Do not rely on this as your only copy of critical data
- **Unaudited cryptography**: Implementation should be audited by security professionals before production use
- **Ephemeral providers**: Anonymous paste services may expire or delete data without notice
- **No guarantees**: Provider availability and data persistence are not guaranteed
- **Maintain backups**: Always keep additional copies of important files
- **Provider independence**: Umbra is not affiliated with any paste service providers — users must comply with each provider's terms of service

Umbra repurposes paste services for unintended use cases. While the cryptography is sound, the storage layer is inherently temporary and experimental.

Use at your own risk.

## Acknowledgments

**Umbra** — Secure, fragmented, redundant storage for the privacy-conscious.
