import java.math.BigInteger;
import java.util.Scanner;

public class FibBigint {

    // Recursive Fibonacci (int)
    public static long fibInt(int n) {
        if (n <= 1) return n;
        return fibInt(n - 1) + fibInt(n - 2);
    }

    // Iterative Fibonacci (long)
    public static long fibIntIter(int n) {
        if (n <= 1) return n;
        long a = 0, b = 1;
        for (int i = 2; i <= n; i++) {
            long temp = a + b;
            a = b;
            b = temp;
        }
        return b;
    }

    // Iterative BigInteger Fibonacci
    public static BigInteger fibBigIntIter(int n) {
        if (n <= 1) return BigInteger.valueOf(n);
        BigInteger a = BigInteger.ZERO;
        BigInteger b = BigInteger.ONE;
        for (int i = 2; i <= n; i++) {
            BigInteger temp = a.add(b);
            a = b;
            b = temp;
        }
        return b;
    }

    public static void main(String[] args) {
        System.out.println("=== Java Fibonacci Benchmark ===\n");

        long start, elapsed;

        // Int recursive
        System.out.println("--- Int (recursive) ---");
        start = System.currentTimeMillis();
        long result = fibInt(35);
        elapsed = System.currentTimeMillis() - start;
        System.out.printf("fib(35) = %d | time: %d ms%n", result, elapsed);

        start = System.currentTimeMillis();
        result = fibInt(40);
        elapsed = System.currentTimeMillis() - start;
        System.out.printf("fib(40) = %d | time: %d ms%n", result, elapsed);

        // Int iterative
        System.out.println("\n--- Int (iterative) ---");
        start = System.currentTimeMillis();
        result = fibIntIter(90);
        elapsed = System.currentTimeMillis() - start;
        System.out.printf("fib(90) = %d | time: %d ms%n", result, elapsed);

        // BigInt iterative
        System.out.println("\n--- BigInt (iterative) ---");
        start = System.currentTimeMillis();
        BigInteger bigResult = fibBigIntIter(100);
        elapsed = System.currentTimeMillis() - start;
        System.out.printf("fib(100) bits: %d | time: %d ms%n", bigResult.bitLength(), elapsed);

        start = System.currentTimeMillis();
        bigResult = fibBigIntIter(1000);
        elapsed = System.currentTimeMillis() - start;
        System.out.printf("fib(1000) bits: %d | time: %d ms%n", bigResult.bitLength(), elapsed);

        start = System.currentTimeMillis();
        bigResult = fibBigIntIter(10000);
        elapsed = System.currentTimeMillis() - start;
        System.out.printf("fib(10000) bits: %d | time: %d ms%n", bigResult.bitLength(), elapsed);

        start = System.currentTimeMillis();
        bigResult = fibBigIntIter(100000);
        elapsed = System.currentTimeMillis() - start;
        System.out.printf("fib(100000) bits: %d | time: %d ms%n", bigResult.bitLength(), elapsed);

        String str = bigResult.toString();
        System.out.printf("fib(100000) length: %d digits%n", str.length());
        System.out.printf("fib(100000) first 50: %s%n", str.substring(0, Math.min(50, str.length())));
        System.out.printf("fib(100000) last 50: %s%n", str.substring(Math.max(0, str.length() - 50)));
    }
}
