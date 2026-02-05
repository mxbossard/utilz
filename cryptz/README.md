## Purpose
An abstraction to os.File to encrypt data without having to manage encryption.

## File name
File names are composed of:
- public path
- private path

The public path is a dir path which is not tempered.

By default, the private path is a dir path plus a file path which will be hashed. 
Each part of the private path is hashed individually.
Ex: /publicDir1/publicDir2/privateDir1/privateDir2/filename will be transformed to: /publicDir1/publicDir2/<hash1>/<hash2>/<hash3>

It is possible to keep file names in plain text.

## File Content
The file content is buffered in memory until Flush(), Sync() or Close() methode are used.
When written on disk, the file content is ciphered and never happen to be plaintext on disk.
Two identical contents are not ciphered the same for security reason (nonce is not reused).
Not modyfing the file contents does not trigger a reciphering.
The nonce used to cipher the content is public and written on the file head.

## Security considerations
- Key is secret.
- Nonce MUST be unique and used once for an encryption.
- Keys and Hashing functions are salted.
- Always use at least 2 ciphering algorythm sequentially to increase security.