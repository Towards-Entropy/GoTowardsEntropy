import os

def split_file(file_path, chunk_size=10240):
    # Check if file exists
    if not os.path.exists(file_path):
        print("File not found.")
        return
    
    file_dir, file_name = os.path.split(file_path)
    base_name, ext = os.path.splitext(file_name)
    
    # Open file for reading
    with open(file_path, 'rb') as f:
        chunk_count = 0
        while True:
            # Read only 10KB
            chunk = f.read(chunk_size)
            
            if not chunk:  # End of file
                break
            
            # Prepare output file name
            chunk_file_name = f"{base_name}_chunk_{chunk_count}{ext}"
            chunk_file_path = os.path.join(file_dir, chunk_file_name)
            
            # Write the chunk to a new file
            with open(chunk_file_path, 'wb') as chunk_file:
                chunk_file.write(chunk)
            
            print(f"Wrote {chunk_file_path}")
            
            chunk_count += 1

if __name__ == "__main__":
    file_to_split = input("Enter the name of the file to split: ")
    split_file(file_to_split)
