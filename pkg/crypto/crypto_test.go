package crypto

import (
	"bytes"
	"testing"
)

func TestVerifyPassword(t *testing.T) {
	type args struct {
		password string
		salt     []byte
		hash     []byte
		iter     int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				password: "a/XfkAUZTnzKgLeLa7e7PsJURVDAxgRJXVUIiJOI5cU=",
				salt:     []byte{50, 16, 172, 0, 53, 22, 73, 230, 32, 32, 16, 29, 136, 71, 245, 66, 35, 63, 80, 143, 220, 71, 241, 159, 61, 226, 128, 103, 122, 135, 207, 223, 25, 223, 42, 176, 121, 248, 184, 148, 29, 146, 141, 39, 211, 203, 151, 132, 236, 1, 114, 196, 10, 111, 45, 78, 136, 193, 55, 98, 194, 83, 38, 106},
				hash:     []byte{174, 106, 48, 161, 162, 97, 167, 103, 0, 131, 59, 156, 68, 241, 155, 208, 80, 8, 245, 2, 206, 209, 196, 163, 148, 186, 138, 175, 61, 15, 238, 39},
				iter:     100000,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := VerifyPassword(
				tt.args.password,
				tt.args.salt,
				tt.args.hash,
				tt.args.iter); got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGeneratePassword(t *testing.T) {
	type args struct {
		password string
		salt     []byte
		iter     int
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GeneratePassword(
				tt.args.password,
				tt.args.salt,
				tt.args.iter,
			)
			if bytes.Equal(got, tt.want) {
				t.Errorf("GeneratePassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenAndVerifyPassword(t *testing.T) {
	type args struct {
		password string
		salt     []byte
		hash     []byte
		iter     int
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "test1",
			args: args{
				password: "7IlPno0jtNjyVR3zOlPjL6gAU2/k/dFeOHzwxt9vqG4=",
				salt:     []byte{77, 231, 60, 223, 221, 62, 51, 170, 126, 148, 79, 249, 148, 3, 21, 89, 112, 112, 47, 223, 96, 120, 186, 224, 186, 120, 223, 140, 189, 240, 182, 170, 101, 22, 70, 208, 219, 129, 220, 118, 76, 162, 176, 231, 239, 157, 132, 52, 79, 166, 147, 38, 116, 16, 179, 9, 41, 141, 119, 185, 154, 28, 48, 1},
				hash:     []byte{192, 173, 209, 229, 146, 127, 46, 202, 150, 158, 178, 120, 92, 158, 221, 170, 110, 242, 100, 143, 35, 143, 60, 38, 38, 63, 118, 231, 64, 61, 146, 118},
				iter:     100000,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash := GeneratePassword(tt.args.password, tt.args.salt, tt.args.iter)
			if !bytes.Equal(hash, tt.args.hash) {
				t.Errorf("GeneratePassword() = %v, want %v", hash, tt.args.hash)
			}

			if !VerifyPassword(tt.args.password, tt.args.salt, hash, tt.args.iter) {
				t.Errorf("VerifyPassword() error")
				return
			}
		})
	}
}
