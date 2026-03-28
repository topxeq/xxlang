#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

// Simple BigInt implementation for C benchmark
typedef struct {
    unsigned int *digits;
    int len;
    int capacity;
} BigInt;

BigInt* bigint_new(int capacity) {
    BigInt *b = malloc(sizeof(BigInt));
    b->digits = calloc(capacity, sizeof(unsigned int));
    b->len = 0;
    b->capacity = capacity;
    return b;
}

void bigint_free(BigInt *b) {
    free(b->digits);
    free(b);
}

BigInt* bigint_from_int(unsigned int n) {
    BigInt *b = bigint_new(1);
    if (n > 0) {
        b->digits[0] = n;
        b->len = 1;
    }
    return b;
}

BigInt* bigint_copy(BigInt *a) {
    BigInt *b = bigint_new(a->capacity);
    b->len = a->len;
    memcpy(b->digits, a->digits, a->len * sizeof(unsigned int));
    return b;
}

// Add two BigInts, result stored in a
void bigint_add(BigInt *a, BigInt *b) {
    unsigned long long carry = 0;
    int maxLen = (a->len > b->len) ? a->len : b->len;

    // Resize if needed
    if (maxLen + 1 > a->capacity) {
        int newCap = maxLen + 1;
        a->digits = realloc(a->digits, newCap * sizeof(unsigned int));
        memset(a->digits + a->capacity, 0, (newCap - a->capacity) * sizeof(unsigned int));
        a->capacity = newCap;
    }

    for (int i = 0; i < maxLen || carry; i++) {
        unsigned long long sum = carry;
        if (i < a->len) sum += a->digits[i];
        if (i < b->len) sum += b->digits[i];

        a->digits[i] = (unsigned int)(sum & 0xFFFFFFFF);
        carry = sum >> 32;

        if (i >= a->len) a->len = i + 1;
    }
}

int bigint_bitlen(BigInt *a) {
    if (a->len == 0) return 0;
    int bits = (a->len - 1) * 32;
    unsigned int top = a->digits[a->len - 1];
    while (top) {
        bits++;
        top >>= 1;
    }
    return bits;
}

// Recursive Fibonacci (int)
long long fib_int(int n) {
    if (n <= 1) return n;
    return fib_int(n - 1) + fib_int(n - 2);
}

// Iterative Fibonacci (long long)
long long fib_int_iter(int n) {
    if (n <= 1) return n;
    long long a = 0, b = 1;
    for (int i = 2; i <= n; i++) {
        long long temp = a + b;
        a = b;
        b = temp;
    }
    return b;
}

// Iterative BigInt Fibonacci
BigInt* fib_bigint_iter(int n) {
    if (n <= 1) return bigint_from_int(n);

    BigInt *a = bigint_from_int(0);
    BigInt *b = bigint_from_int(1);

    for (int i = 2; i <= n; i++) {
        bigint_add(a, b);  // a = a + b
        // swap a and b
        BigInt *temp = a;
        a = b;
        b = temp;
    }

    bigint_free(a);
    return b;
}

int main() {
    printf("=== C Fibonacci Benchmark ===\n\n");

    clock_t start, end;
    long long result;
    BigInt *bigResult;

    // Int recursive
    printf("--- Int (recursive) ---\n");
    start = clock();
    result = fib_int(35);
    end = clock();
    printf("fib(35) = %lld | time: %ld ms\n", result, (end - start) * 1000 / CLOCKS_PER_SEC);

    start = clock();
    result = fib_int(40);
    end = clock();
    printf("fib(40) = %lld | time: %ld ms\n", result, (end - start) * 1000 / CLOCKS_PER_SEC);

    // Int iterative
    printf("\n--- Int (iterative) ---\n");
    start = clock();
    result = fib_int_iter(90);
    end = clock();
    printf("fib(90) = %lld | time: %ld ms\n", result, (end - start) * 1000 / CLOCKS_PER_SEC);

    // BigInt iterative
    printf("\n--- BigInt (iterative) ---\n");
    start = clock();
    bigResult = fib_bigint_iter(100);
    end = clock();
    printf("fib(100) bits: %d | time: %ld ms\n", bigint_bitlen(bigResult), (end - start) * 1000 / CLOCKS_PER_SEC);
    bigint_free(bigResult);

    start = clock();
    bigResult = fib_bigint_iter(1000);
    end = clock();
    printf("fib(1000) bits: %d | time: %ld ms\n", bigint_bitlen(bigResult), (end - start) * 1000 / CLOCKS_PER_SEC);
    bigint_free(bigResult);

    start = clock();
    bigResult = fib_bigint_iter(10000);
    end = clock();
    printf("fib(10000) bits: %d | time: %ld ms\n", bigint_bitlen(bigResult), (end - start) * 1000 / CLOCKS_PER_SEC);
    bigint_free(bigResult);

    start = clock();
    bigResult = fib_bigint_iter(100000);
    end = clock();
    printf("fib(100000) bits: %d | time: %ld ms\n", bigint_bitlen(bigResult), (end - start) * 1000 / CLOCKS_PER_SEC);
    bigint_free(bigResult);

    return 0;
}
