package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Struct utama untuk data koleksi museum
type Koleksi struct {
	ID                primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Kategori          Kategori           `json:"kategori,omitempty" bson:"kategori,omitempty"`
	NoRegistrasi      string             `json:"no_reg,omitempty" bson:"no_reg,omitempty"`
	NoInventaris      string             `json:"no_inv,omitempty" bson:"no_inv,omitempty"`
	NamaBenda         string             `json:"nama_benda,omitempty" bson:"nama_benda,omitempty"`
	AsalKoleksi       string             `json:"asal_koleksi,omitempty" bson:"asal_koleksi,omitempty"`
	Bahan             string             `json:"bahan,omitempty" bson:"bahan,omitempty"`
	Ukuran            Ukuran             `json:"ukuran,omitempty" bson:"ukuran,omitempty"`
	TempatPerolehan   string             `json:"tempat_perolehan,omitempty" bson:"tempat_perolehan,omitempty"`
	TanggalPerolehan  string             `json:"tanggal_perolehan,omitempty" bson:"tanggal_perolehan,omitempty"`
	Deskripsi         string             `json:"deskripsi,omitempty" bson:"deskripsi,omitempty"`
	TempatPenyimpanan string             `json:"tempat_penyimpanan,omitempty" bson:"tempat_penyimpanan,omitempty"`
	Kondisi           string             `json:"kondisi,omitempty" bson:"kondisi,omitempty"`
	Foto              string             `json:"foto,omitempty" bson:"foto,omitempty"`
	CreatedAt         time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
}

type Ukuran struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id"`
	Lebar              string             `json:"lebar,omitempty" bson:"lebar,omitempty"`
	Tebal              string             `json:"tebal,omitempty" bson:"tebal,omitempty"`
	Tinggi             string             `json:"tinggi,omitempty" bson:"tinggi,omitempty"`
	Diameter           string             `json:"diameter,omitempty" bson:"diameter,omitempty"`
	Berat              string             `json:"berat,omitempty" bson:"berat,omitempty"`
	PanjangKeseluruhan string             `json:"panjang_keseluruhan,omitempty" bson:"panjang_keseluruhan,omitempty"`
	CreatedAt          time.Time          `json:"created_at,omitempty" bson:"created_at,omitempty"`
}
