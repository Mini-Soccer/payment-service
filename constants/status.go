package constants

type PaymentStatus int
type PaymentStatusString string

/*
- Initial => order baru dibuat, midtrans belum dipanggil, dan belum ada transaksi pembayaran. misal, user klik checkout tapi belum diarahkan ke midtrans
- Pending => transaksi sudah dibuat di midtrans, user belum menyelesaikan pembayaran. misal, user dapat VA / QRIS / Redirect Link, tapi belum bayar
- Settlement => pembayaran berhasil, dana sudah diterima, transaksi dianggap final. misal, user bayar -> sukses -> order boleh diproses
- Expire => waktu pembayaran habis, transaksi gagal otomatis, user harus buat transaksi baru. misal, VA 24 jam tidak dibayar -> expired
*/

const (
	Initial    PaymentStatus = 0
	Pending    PaymentStatus = 100
	Settlement PaymentStatus = 200
	Expire     PaymentStatus = 300

	InitialString    PaymentStatusString = "initial"
	PendingString    PaymentStatusString = "pending"
	SettlementString PaymentStatusString = "settlement"
	ExpireString     PaymentStatusString = "expire"
)

var mapStatusStringToInt = map[PaymentStatusString]PaymentStatus{
	InitialString:    Initial,
	PendingString:    Pending,
	SettlementString: Settlement,
	ExpireString:     Expire,
}

var mapStatusIntToString = map[PaymentStatus]PaymentStatusString{
	Initial:    InitialString,
	Pending:    PendingString,
	Settlement: SettlementString,
	Expire:     ExpireString,
}

func (p PaymentStatusString) String() string {
	return string(p)
}

func (p PaymentStatus) Int() int {
	return int(p)
}

func (p PaymentStatus) GetStatusString() PaymentStatusString {
	return mapStatusIntToString[p]
}

func (p PaymentStatusString) GetStatusInt() PaymentStatus {
	return mapStatusStringToInt[p]
}
