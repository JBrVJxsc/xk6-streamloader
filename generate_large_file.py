import os
import random
import string

def generate_large_file(file_path, size_in_mb):
    """
    Generates a large file with random text content.

    :param file_path: The path to the file to be created.
    :param size_in_mb: The desired file size in megabytes.
    """
    size_in_bytes = size_in_mb * 1024 * 1024
    chunk_size = 1024  # 1 KB chunks
    
    # Generate a random chunk of text
    random_text_chunk = ''.join(random.choices(string.ascii_letters + string.digits + ' \n', k=chunk_size))
    
    with open(file_path, 'w') as f:
        bytes_written = 0
        while bytes_written < size_in_bytes:
            f.write(random_text_chunk)
            bytes_written += len(random_text_chunk.encode('utf-8'))
            if bytes_written + chunk_size > size_in_bytes:
                # Adjust the last chunk to get the exact file size
                remaining_bytes = size_in_bytes - bytes_written
                if remaining_bytes > 0:
                    last_chunk = ''.join(random.choices(string.ascii_letters + string.digits + ' ', k=remaining_bytes))
                    f.write(last_chunk)
                break

    print(f"Successfully generated '{file_path}' with size {os.path.getsize(file_path) / (1024 * 1024):.2f} MB")

if __name__ == "__main__":
    file_name = "large_file.txt"
    file_size_mb = 50
    generate_large_file(file_name, file_size_mb) 