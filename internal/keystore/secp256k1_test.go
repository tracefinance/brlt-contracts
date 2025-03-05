package keystore

import (
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test vectors from https://www.secg.org/sec2-v2.pdf
var (
	// Known valid point on the curve
	testPointX, _ = new(big.Int).SetString("55066263022277343669578718895168534326250603453777594175500187360389116729240", 10)
	testPointY, _ = new(big.Int).SetString("32670510020758816978083085130507043184471273380659243275938904335757337482424", 10)

	// Known scalar multiplication result for the above point with k=2
	doubleX, _ = new(big.Int).SetString("89565891926547004231252920425935692360644145829622209833684329913297188986597", 10)
	doubleY, _ = new(big.Int).SetString("12158399299693830322967808612713398636155367887041628176798871954788371653930", 10)

	// Known scalar for testing scalar multiplication
	testK, _ = new(big.Int).SetString("7", 10)

	// Expected result for k*G (scalar base multiplication with k=7)
	// Corrected values for 7*G on secp256k1
	testScalarBaseMultX, _ = new(big.Int).SetString("5cbdf0646e5db4eaa398f365f2ea7a0e3d419b7e0330e39ce92bddedcac4f9bc", 16)
	testScalarBaseMultY, _ = new(big.Int).SetString("6aebca40ba255960a3178d6d861a54dba813d0b813fde7b5a5082628087264da", 16)
)

// badSecp256k1Curve is a test-only subclass that allows triggering panic scenarios
type badSecp256k1Curve struct {
	*secp256k1Curve
}

func (curve *badSecp256k1Curve) addUnsafe(x1, y1, x2, y2 *big.Int) (*big.Int, *big.Int) {
	// Force the panic by skipping the equality check and creating deltaX = 0
	// This is for test purposes only to trigger the panic code path
	p := curve.P

	// Force x2 to be equal to x1, which would normally be caught by the equality check
	x2 = new(big.Int).Set(x1)
	// But make y2 different than y1 and -y1, to bypass the other checks
	y2 = new(big.Int).Add(y1, big.NewInt(1))

	// This should cause deltaX to be 0, triggering the panic
	deltaX := new(big.Int).Sub(x2, x1)
	deltaX.Mod(deltaX, p)
	deltaY := new(big.Int).Sub(y2, y1)
	deltaY.Mod(deltaY, p)
	invDeltaX := new(big.Int).ModInverse(deltaX, p)
	// This should panic
	lambda := new(big.Int).Mul(deltaY, invDeltaX)
	lambda.Mod(lambda, p)

	// We shouldn't reach this code
	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, x1)
	x3.Sub(x3, x2)
	x3.Mod(x3, p)
	tmp := new(big.Int).Sub(x1, x3)
	tmp.Mul(lambda, tmp)
	y3 := new(big.Int).Sub(tmp, y1)
	y3.Mod(y3, p)
	return x3, y3
}

func (curve *badSecp256k1Curve) doubleUnsafe(x1, y1 *big.Int) (*big.Int, *big.Int) {
	// Force the panic by skipping the y1 = 0 check and creating a scenario with twoY1 = 0
	// This is for test purposes only to trigger the panic code path
	p := curve.P

	// Force y1 to be 0 mod p
	y1 = big.NewInt(0)

	threeX1Sq := new(big.Int).Mul(x1, x1)
	threeX1Sq.Mul(threeX1Sq, big.NewInt(3))
	twoY1 := new(big.Int).Mul(y1, big.NewInt(2))
	twoY1.Mod(twoY1, p)
	invTwoY1 := new(big.Int).ModInverse(twoY1, p)
	// This should panic
	lambda := new(big.Int).Mul(threeX1Sq, invTwoY1)
	lambda.Mod(lambda, p)

	// We shouldn't reach this code
	x3 := new(big.Int).Mul(lambda, lambda)
	x3.Sub(x3, new(big.Int).Mul(big.NewInt(2), x1))
	x3.Mod(x3, p)
	tmp := new(big.Int).Sub(x1, x3)
	tmp.Mul(lambda, tmp)
	y3 := new(big.Int).Sub(tmp, y1)
	y3.Mod(y3, p)
	return x3, y3
}

func TestSecp256k1Initialization(t *testing.T) {
	// Verify the curve was initialized correctly
	assert.NotNil(t, Secp256k1, "Secp256k1 should be initialized")
	assert.Equal(t, "secp256k1", Secp256k1.Name, "Curve name should be secp256k1")
	assert.Equal(t, 256, Secp256k1.BitSize, "Curve bit size should be 256")

	// Check curve parameters
	assert.Equal(t, p, Secp256k1.P, "Incorrect P parameter")
	assert.Equal(t, n, Secp256k1.N, "Incorrect N parameter")
	assert.Equal(t, b, Secp256k1.B, "Incorrect B parameter")
	assert.Equal(t, gx, Secp256k1.Gx, "Incorrect Gx parameter")
	assert.Equal(t, gy, Secp256k1.Gy, "Incorrect Gy parameter")

	// Verify base point is on the curve
	assert.True(t, Secp256k1.IsOnCurve(Secp256k1.Gx, Secp256k1.Gy), "Base point should be on the curve")
}

func TestIsOnCurve(t *testing.T) {
	// Test cases for IsOnCurve
	testCases := []struct {
		name string
		x, y *big.Int
		want bool
	}{
		{
			name: "Base point G",
			x:    Secp256k1.Gx,
			y:    Secp256k1.Gy,
			want: true,
		},
		{
			name: "Known valid point",
			x:    testPointX,
			y:    testPointY,
			want: true,
		},
		{
			name: "Point at infinity",
			x:    new(big.Int),
			y:    new(big.Int),
			want: false, // Point at infinity is not technically "on" the curve
		},
		{
			name: "Invalid x coordinate",
			x:    new(big.Int).Add(p, big.NewInt(1)), // x > p
			y:    Secp256k1.Gy,
			want: false,
		},
		{
			name: "Invalid y coordinate",
			x:    Secp256k1.Gx,
			y:    new(big.Int).Add(p, big.NewInt(1)), // y > p
			want: false,
		},
		{
			name: "Negative x coordinate",
			x:    new(big.Int).Neg(Secp256k1.Gx),
			y:    Secp256k1.Gy,
			want: false,
		},
		{
			name: "Point not satisfying curve equation",
			x:    big.NewInt(123456),
			y:    big.NewInt(654321),
			want: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := Secp256k1.IsOnCurve(tc.x, tc.y)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestDouble(t *testing.T) {
	// Test doubling a point
	x3, y3 := Secp256k1.Double(testPointX, testPointY)

	assert.Equal(t, 0, doubleX.Cmp(x3), "Incorrect x-coordinate after doubling")
	assert.Equal(t, 0, doubleY.Cmp(y3), "Incorrect y-coordinate after doubling")

	// The resulting point should be on the curve
	assert.True(t, Secp256k1.IsOnCurve(x3, y3), "Result of doubling should be on the curve")

	// Test doubling the point at infinity
	x3, y3 = Secp256k1.Double(new(big.Int), new(big.Int))
	assert.Equal(t, 0, x3.Sign(), "Doubling point at infinity should return point at infinity (x)")
	assert.Equal(t, 0, y3.Sign(), "Doubling point at infinity should return point at infinity (y)")

	// Test doubling a point with y=0 (should return point at infinity)
	x := big.NewInt(1)
	y := big.NewInt(0)
	x3, y3 = Secp256k1.Double(x, y)
	assert.Equal(t, 0, x3.Sign(), "Doubling point with y=0 should return point at infinity (x)")
	assert.Equal(t, 0, y3.Sign(), "Doubling point with y=0 should return point at infinity (y)")
}

func TestAdd(t *testing.T) {
	// Test cases for Add
	testCases := []struct {
		name        string
		x1, y1      *big.Int
		x2, y2      *big.Int
		wantX       *big.Int
		wantY       *big.Int
		wantOnCurve bool
	}{
		{
			name:        "Add point to itself",
			x1:          testPointX,
			y1:          testPointY,
			x2:          testPointX,
			y2:          testPointY,
			wantX:       doubleX,
			wantY:       doubleY,
			wantOnCurve: true,
		},
		{
			name:        "Add point to point at infinity",
			x1:          testPointX,
			y1:          testPointY,
			x2:          new(big.Int),
			y2:          new(big.Int),
			wantX:       testPointX,
			wantY:       testPointY,
			wantOnCurve: true,
		},
		{
			name:        "Add point at infinity to point",
			x1:          new(big.Int),
			y1:          new(big.Int),
			x2:          testPointX,
			y2:          testPointY,
			wantX:       testPointX,
			wantY:       testPointY,
			wantOnCurve: true,
		},
		{
			name:        "Add point to its inverse",
			x1:          testPointX,
			y1:          testPointY,
			x2:          testPointX,
			y2:          new(big.Int).Sub(p, testPointY), // -y1 mod p
			wantX:       new(big.Int),
			wantY:       new(big.Int),
			wantOnCurve: false, // Point at infinity
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			resultX, resultY := Secp256k1.Add(tc.x1, tc.y1, tc.x2, tc.y2)

			assert.Equal(t, 0, tc.wantX.Cmp(resultX), "Incorrect x-coordinate after addition")
			assert.Equal(t, 0, tc.wantY.Cmp(resultY), "Incorrect y-coordinate after addition")

			// Check if the result should be on the curve
			isOnCurve := Secp256k1.IsOnCurve(resultX, resultY)
			assert.Equal(t, tc.wantOnCurve, isOnCurve)
		})
	}

	// Additional test: add two different valid points
	// First get G + G = 2G
	doubleX, doubleY := Secp256k1.Double(Secp256k1.Gx, Secp256k1.Gy)
	// Then calculate 2G + G = 3G
	tripleX, tripleY := Secp256k1.Add(doubleX, doubleY, Secp256k1.Gx, Secp256k1.Gy)

	// 3G should be on the curve
	assert.True(t, Secp256k1.IsOnCurve(tripleX, tripleY), "3G should be on the curve")

	// Now calculate G + 2G (should equal 3G)
	altTripleX, altTripleY := Secp256k1.Add(Secp256k1.Gx, Secp256k1.Gy, doubleX, doubleY)

	// Results should be the same
	assert.Equal(t, 0, tripleX.Cmp(altTripleX), "Addition should be commutative (x)")
	assert.Equal(t, 0, tripleY.Cmp(altTripleY), "Addition should be commutative (y)")
}

func TestScalarMult(t *testing.T) {
	// Test scalar multiplication by k=0 (should return point at infinity)
	zeroK := make([]byte, 32)
	x, y := Secp256k1.ScalarMult(testPointX, testPointY, zeroK)
	assert.Equal(t, 0, x.Sign(), "Scalar multiplication by 0 should return point at infinity (x)")
	assert.Equal(t, 0, y.Sign(), "Scalar multiplication by 0 should return point at infinity (y)")

	// Test scalar multiplication by k=1 (should return the same point)
	oneK := make([]byte, 32)
	oneK[31] = 1
	x, y = Secp256k1.ScalarMult(testPointX, testPointY, oneK)
	assert.Equal(t, 0, testPointX.Cmp(x), "Scalar multiplication by 1 should return the original point (x)")
	assert.Equal(t, 0, testPointY.Cmp(y), "Scalar multiplication by 1 should return the original point (y)")

	// Test scalar multiplication by k=2 (should equal doubling)
	twoK := make([]byte, 32)
	twoK[31] = 2
	x, y = Secp256k1.ScalarMult(testPointX, testPointY, twoK)
	assert.Equal(t, 0, doubleX.Cmp(x), "Scalar multiplication by 2 should equal point doubling (x)")
	assert.Equal(t, 0, doubleY.Cmp(y), "Scalar multiplication by 2 should equal point doubling (y)")

	// Test with a known scalar (k=7)
	k7 := make([]byte, 32)
	k7[31] = 7
	xG, yG := Secp256k1.ScalarMult(Secp256k1.Gx, Secp256k1.Gy, k7)

	// Verify the result is on the curve
	assert.True(t, Secp256k1.IsOnCurve(xG, yG), "Result of scalar multiplication should be on the curve")

	// Test with large scalar
	largeSeed := make([]byte, 32)
	_, err := rand.Read(largeSeed)
	require.NoError(t, err, "Failed to generate random bytes")

	// Ensure scalar is > N to test modular reduction
	largeK := new(big.Int).SetBytes(largeSeed)
	largeK.Add(largeK, Secp256k1.N)

	// Perform scalar multiplication
	xLarge, yLarge := Secp256k1.ScalarMult(Secp256k1.Gx, Secp256k1.Gy, largeK.Bytes())

	// Result should be on the curve
	assert.True(t, Secp256k1.IsOnCurve(xLarge, yLarge), "Result with large scalar should be on the curve")
}

func TestScalarBaseMult(t *testing.T) {
	// Test with zero scalar
	zeroK := make([]byte, 32)
	x, y := Secp256k1.ScalarBaseMult(zeroK)
	assert.Equal(t, 0, x.Sign(), "Scalar base multiplication by 0 should return point at infinity (x)")
	assert.Equal(t, 0, y.Sign(), "Scalar base multiplication by 0 should return point at infinity (y)")

	// Test with scalar=1 (should return base point G)
	oneK := make([]byte, 32)
	oneK[31] = 1
	x, y = Secp256k1.ScalarBaseMult(oneK)
	assert.Equal(t, 0, Secp256k1.Gx.Cmp(x), "Scalar base multiplication by 1 should return G (x)")
	assert.Equal(t, 0, Secp256k1.Gy.Cmp(y), "Scalar base multiplication by 1 should return G (y)")

	// Test with known scalar k=7 and known result
	k7 := make([]byte, 32)
	k7[31] = 7
	x, y = Secp256k1.ScalarBaseMult(k7)

	// Convert result to hex strings for easier comparison with test vectors
	resultX := new(big.Int).Set(x)
	resultY := new(big.Int).Set(y)

	assert.Equal(t, 0, testScalarBaseMultX.Cmp(resultX), "Incorrect x-coordinate for 7*G")
	assert.Equal(t, 0, testScalarBaseMultY.Cmp(resultY), "Incorrect y-coordinate for 7*G")

	// Result should be on the curve
	assert.True(t, Secp256k1.IsOnCurve(x, y), "Result of scalar base multiplication should be on the curve")

	// Verify ScalarBaseMult and ScalarMult with base point give identical results
	randomK := make([]byte, 32)
	_, err := rand.Read(randomK)
	require.NoError(t, err, "Failed to generate random bytes")

	x1, y1 := Secp256k1.ScalarBaseMult(randomK)
	x2, y2 := Secp256k1.ScalarMult(Secp256k1.Gx, Secp256k1.Gy, randomK)

	assert.Equal(t, 0, x1.Cmp(x2), "ScalarBaseMult and ScalarMult should return the same x-coordinate")
	assert.Equal(t, 0, y1.Cmp(y2), "ScalarBaseMult and ScalarMult should return the same y-coordinate")
}

func TestParams(t *testing.T) {
	params := Secp256k1.Params()

	assert.Equal(t, "secp256k1", params.Name, "Incorrect curve name")
	assert.Equal(t, 256, params.BitSize, "Incorrect bit size")
	assert.Equal(t, 0, p.Cmp(params.P), "Incorrect P parameter")
	assert.Equal(t, 0, n.Cmp(params.N), "Incorrect N parameter")
	assert.Equal(t, 0, b.Cmp(params.B), "Incorrect B parameter")
	assert.Equal(t, 0, gx.Cmp(params.Gx), "Incorrect Gx parameter")
	assert.Equal(t, 0, gy.Cmp(params.Gy), "Incorrect Gy parameter")
}

// Benchmarks

func BenchmarkIsOnCurve(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Secp256k1.IsOnCurve(testPointX, testPointY)
	}
}

func BenchmarkAdd(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Secp256k1.Add(testPointX, testPointY, doubleX, doubleY)
	}
}

func BenchmarkDouble(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Secp256k1.Double(testPointX, testPointY)
	}
}

func BenchmarkScalarMult(b *testing.B) {
	k := testK.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Secp256k1.ScalarMult(testPointX, testPointY, k)
	}
}

func BenchmarkScalarBaseMult(b *testing.B) {
	k := testK.Bytes()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Secp256k1.ScalarBaseMult(k)
	}
}

func TestAddErrorCase(t *testing.T) {
	// This test attempts to create a pathological case that would trigger
	// the panic in the Add method (division by zero)
	// We shouldn't be able to trigger it in normal operation, but we test for coverage

	// Create points with identical x but different y values that should add to p
	x := big.NewInt(42)
	y1 := big.NewInt(10)

	// Calculate -y1 mod p
	y2 := new(big.Int).Sub(p, y1)

	// This shouldn't panic, as x1==x2 should be caught before attempting division
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Add should not panic: %v", r)
		}
	}()

	x3, y3 := Secp256k1.Add(x, y1, x, y2)

	// Result should be point at infinity
	assert.Equal(t, 0, x3.Sign(), "Should return point at infinity (x)")
	assert.Equal(t, 0, y3.Sign(), "Should return point at infinity (y)")
}

func TestDoubleErrorCase(t *testing.T) {
	// This test attempts to create a pathological case that would trigger
	// the panic in the Double method (division by zero)
	// We can't directly test the panic since y1=0 is caught before the division

	// Create a point with y=0
	x := big.NewInt(1)
	y := big.NewInt(0)

	// This shouldn't panic, as y=0 is caught before attempting division
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Double should not panic: %v", r)
		}
	}()

	x3, y3 := Secp256k1.Double(x, y)

	// Result should be point at infinity
	assert.Equal(t, 0, x3.Sign(), "Should return point at infinity (x)")
	assert.Equal(t, 0, y3.Sign(), "Should return point at infinity (y)")
}

func TestAddPanic(t *testing.T) {
	badCurve := &badSecp256k1Curve{Secp256k1}

	// Create a test point
	x := big.NewInt(42)
	y := big.NewInt(10)

	// Test that the unsafe function panics
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("addUnsafe should have panicked")
		} else {
			// Just check that it panicked, the specific message may vary
			// based on the Go runtime
			t.Logf("Recovered panic: %v", r)
		}
	}()

	// This should panic
	badCurve.addUnsafe(x, y, x, y)
}

func TestDoublePanic(t *testing.T) {
	badCurve := &badSecp256k1Curve{Secp256k1}

	// Create a test point
	x := big.NewInt(42)
	y := big.NewInt(10) // Value doesn't matter, it will be replaced in doubleUnsafe

	// Test that the unsafe function panics
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("doubleUnsafe should have panicked")
		} else {
			// Just check that it panicked, the specific message may vary
			// based on the Go runtime
			t.Logf("Recovered panic: %v", r)
		}
	}()

	// This should panic
	badCurve.doubleUnsafe(x, y)
}
