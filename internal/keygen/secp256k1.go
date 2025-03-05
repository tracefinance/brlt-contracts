package keygen

import (
	"crypto/elliptic"
	"math/big"
)

// secp256k1Curve represents the secp256k1 elliptic curve.
type secp256k1Curve struct {
	*elliptic.CurveParams
}

// Predefined curve parameters as big.Int values.
var (
	// P is the prime defining the finite field: 2^256 - 2^32 - 977
	p = new(big.Int).SetBytes([]byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFE, 0xFF, 0xFF, 0xFC, 0x2F,
	})
	// N is the order of the base point
	n = new(big.Int).SetBytes([]byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
		0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFE,
		0xBA, 0xAE, 0xDC, 0xE6, 0xAF, 0x48, 0xA0, 0x3B,
		0xBF, 0xD2, 0x5E, 0x8C, 0xD0, 0x36, 0x41, 0x41,
	})
	// B is the constant in the curve equation y^2 = x^3 + B
	b = big.NewInt(7)
	// Gx, Gy are the coordinates of the base point G
	gx = new(big.Int).SetBytes([]byte{
		0x79, 0xBE, 0x66, 0x7E, 0xF9, 0xDC, 0xBB, 0xAC,
		0x55, 0xA0, 0x62, 0x95, 0xCE, 0x87, 0x0B, 0x07,
		0x02, 0x9B, 0xFC, 0xDB, 0x2D, 0xCE, 0x28, 0xD9,
		0x59, 0xF2, 0x81, 0x5B, 0x16, 0xF8, 0x17, 0x98,
	})
	gy = new(big.Int).SetBytes([]byte{
		0x48, 0x3A, 0xDA, 0x77, 0x26, 0xA3, 0xC4, 0x65,
		0x5D, 0xA4, 0xFB, 0xFC, 0x0E, 0x11, 0x08, 0xA8,
		0xFD, 0x17, 0xB4, 0x48, 0xA6, 0x85, 0x54, 0x19,
		0x9C, 0x47, 0xD0, 0x8F, 0xFB, 0x10, 0xD4, 0xB8,
	})
)

// Secp256k1 is the global instance of the secp256k1 curve.
var Secp256k1 *secp256k1Curve

// init initializes the Secp256k1 curve with its parameters.
func init() {
	Secp256k1 = &secp256k1Curve{
		CurveParams: &elliptic.CurveParams{
			P:       p,
			N:       n,
			B:       b,
			Gx:      gx,
			Gy:      gy,
			BitSize: 256,
			Name:    "secp256k1",
		},
	}
}

// Params returns the curve parameters.
func (curve *secp256k1Curve) Params() *elliptic.CurveParams {
	return curve.CurveParams
}

// IsOnCurve checks if the point (x, y) lies on the secp256k1 curve.
func (curve *secp256k1Curve) IsOnCurve(x, y *big.Int) bool {
	// Check if coordinates are within the field
	if x.Sign() < 0 || x.Cmp(curve.P) >= 0 || y.Sign() < 0 || y.Cmp(curve.P) >= 0 {
		return false
	}
	// Compute y^2 mod P
	y2 := new(big.Int).Mul(y, y)
	y2.Mod(y2, curve.P)
	// Compute x^3 + B mod P (A = 0 for secp256k1)
	x3 := new(big.Int).Exp(x, big.NewInt(3), curve.P)
	x3.Add(x3, curve.B)
	x3.Mod(x3, curve.P)
	// Point is on curve if y^2 == x^3 + B mod P
	return y2.Cmp(x3) == 0
}

// Add computes the sum of two points (x1, y1) and (x2, y2) on the curve.
func (curve *secp256k1Curve) Add(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	// Handle point at infinity
	if x1.Sign() == 0 && y1.Sign() == 0 {
		return new(big.Int).Set(x2), new(big.Int).Set(y2)
	}
	if x2.Sign() == 0 && y2.Sign() == 0 {
		return new(big.Int).Set(x1), new(big.Int).Set(y1)
	}
	p := curve.P
	// Check if points are inverses (x1 == x2, y1 == -y2 mod P)
	if x1.Cmp(x2) == 0 {
		negY2 := new(big.Int).Neg(y2)
		negY2.Mod(negY2, p)
		if y1.Cmp(negY2) == 0 {
			return new(big.Int), new(big.Int) // Point at infinity
		}
	}
	// Check if points are the same (use Double)
	if x1.Cmp(x2) == 0 && y1.Cmp(y2) == 0 {
		return curve.Double(x1, y1)
	}
	// Compute slope λ = (y2 - y1) / (x2 - x1) mod P
	deltaX := new(big.Int).Sub(x2, x1)
	deltaX.Mod(deltaX, p)
	deltaY := new(big.Int).Sub(y2, y1)
	deltaY.Mod(deltaY, p)
	invDeltaX := new(big.Int).ModInverse(deltaX, p)
	if invDeltaX == nil {
		// This should not occur since x1 != x2 mod P
		panic("division by zero in Add")
	}
	lambda := new(big.Int).Mul(deltaY, invDeltaX)
	lambda.Mod(lambda, p)
	// Compute x3 = λ^2 - x1 - x2 mod P
	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, x1)
	x3.Sub(x3, x2)
	x3.Mod(x3, p)
	// Compute y3 = λ*(x1 - x3) - y1 mod P
	tmp := new(big.Int).Sub(x1, x3)
	tmp.Mul(lambda, tmp)
	y3 := new(big.Int).Sub(tmp, y1)
	y3.Mod(y3, p)
	return x3, y3
}

// Double computes 2*(x1, y1) on the curve.
func (curve *secp256k1Curve) Double(x1, y1 *big.Int) (*big.Int, *big.Int) {
	// If y1 = 0, result is point at infinity
	if y1.Sign() == 0 {
		return new(big.Int), new(big.Int)
	}
	p := curve.P
	// Compute λ = (3*x1^2) / (2*y1) mod P (A = 0 for secp256k1)
	threeX1Sq := new(big.Int).Mul(x1, x1)
	threeX1Sq.Mul(threeX1Sq, big.NewInt(3))
	twoY1 := new(big.Int).Mul(y1, big.NewInt(2))
	twoY1.Mod(twoY1, p)
	invTwoY1 := new(big.Int).ModInverse(twoY1, p)
	if invTwoY1 == nil {
		// This should not occur since y1 != 0
		panic("division by zero in Double")
	}
	lambda := new(big.Int).Mul(threeX1Sq, invTwoY1)
	lambda.Mod(lambda, p)
	// Compute x3 = λ^2 - 2*x1 mod P
	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, new(big.Int).Mul(big.NewInt(2), x1))
	x3.Mod(x3, p)
	// Compute y3 = λ*(x1 - x3) - y1 mod P
	tmp := new(big.Int).Sub(x1, x3)
	tmp.Mul(lambda, tmp)
	y3 := new(big.Int).Sub(tmp, y1)
	y3.Mod(y3, p)
	return x3, y3
}

// ScalarMult computes k * (x1, y1) using the double-and-add algorithm.
func (curve *secp256k1Curve) ScalarMult(x1, y1 *big.Int, k []byte) (*big.Int, *big.Int) {
	scalar := new(big.Int).SetBytes(k)
	// If scalar is 0, return point at infinity
	if scalar.Sign() == 0 {
		return new(big.Int), new(big.Int)
	}
	// Reduce scalar modulo N to prevent timing attacks
	if scalar.Cmp(curve.N) >= 0 {
		scalar.Mod(scalar, curve.N)
	}
	// Initialize result as point at infinity
	rx, ry := new(big.Int), new(big.Int)
	// Point to multiply
	px, py := new(big.Int).Set(x1), new(big.Int).Set(y1)
	// Double-and-add algorithm
	for i := scalar.BitLen() - 1; i >= 0; i-- {
		rx, ry = curve.Double(rx, ry)
		if scalar.Bit(i) == 1 {
			rx, ry = curve.Add(rx, ry, px, py)
		}
	}
	return rx, ry
}

// ScalarBaseMult computes k * G, where G is the base point.
func (curve *secp256k1Curve) ScalarBaseMult(k []byte) (*big.Int, *big.Int) {
	return curve.ScalarMult(curve.Gx, curve.Gy, k)
}
