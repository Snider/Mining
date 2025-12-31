# Supported Mining Algorithms

This guide provides detailed information about the cryptocurrency mining algorithms supported by the Mining Platform.

## Overview

Mining Platform supports multiple Proof-of-Work (PoW) algorithms across CPU and GPU mining. Each algorithm is optimized for specific hardware and cryptocurrencies.

## Algorithm Matrix

| Algorithm | Type | CPU | GPU (OpenCL) | GPU (CUDA) | Primary Coin | Difficulty |
|-----------|------|-----|--------------|------------|--------------|------------|
| RandomX | Memory-hard | ✅ | ✅ | ✅ | Monero (XMR) | Medium |
| KawPow | GPU-optimized | ❌ | ✅ | ✅ | Ravencoin (RVN) | High |
| ETChash | GPU-optimized | ❌ | ✅ | ✅ | Ethereum Classic (ETC) | High |
| ProgPowZ | GPU-optimized | ❌ | ✅ | ✅ | Zano (ZANO) | High |
| Blake3 | CPU/GPU hybrid | ✅ | ✅ | ✅ | Decred (DCR) | Low |
| CryptoNight | Memory-hard | ✅ | ✅ | ✅ | Various | Medium |

## RandomX

### Overview

RandomX is a proof-of-work algorithm optimized for general-purpose CPUs. It is designed to be ASIC-resistant by using random code execution and memory-hard techniques.

### Characteristics

- **Type:** Memory-hard PoW
- **Memory Requirement:** 2GB RAM minimum
- **Optimal Hardware:** Modern CPUs with large cache
- **ASIC Resistance:** High

### Variants

| Variant | Coin | Notes |
|---------|------|-------|
| rx/0 | Monero (XMR) | Primary Monero algorithm |
| rx/wow | Wownero (WOW) | RandomX variant for Wownero |
| rx/arq | ArQmA (ARQ) | ArQmA-specific parameters |
| rx/graft | Graft (GRFT) | Graft Network |

### Configuration

```json
{
  "algo": "rx/0",
  "threads": 4,
  "cpuPriority": 3,
  "hugePages": true,
  "1gb-pages": false,
  "randomx": {
    "mode": "auto",
    "cache": true,
    "dataset": true
  }
}
```

### Performance Tips

**CPU:**
- Use CPUs with large L3 cache (Ryzen, Threadripper, EPYC)
- Enable huge pages for best performance
- Leave 1-2 threads free for system
- Optimal thread count = (CPU cores - 1)

**Hashrate Examples:**
- Intel i5-12400: ~4-5 KH/s
- AMD Ryzen 5 5600X: ~7-8 KH/s
- AMD Ryzen 9 5950X: ~20-22 KH/s
- AMD Threadripper 3990X: ~60-65 KH/s

### Huge Pages Setup

**Linux:**
```bash
sudo sysctl -w vm.nr_hugepages=1280
echo "vm.nr_hugepages=1280" | sudo tee -a /etc/sysctl.conf
```

**Windows:**
Run as Administrator:
```powershell
# Restart required
```

## KawPow

### Overview

KawPow is a derivative of ProgPoW, specifically designed for Ravencoin. It combines random program generation with memory-hard features.

### Characteristics

- **Type:** GPU-optimized PoW
- **Memory Requirement:** 3-4GB VRAM
- **Optimal Hardware:** Modern NVIDIA/AMD GPUs
- **ASIC Resistance:** High

### Configuration

```json
{
  "algo": "kawpow",
  "cuda": {
    "enabled": true,
    "devices": [0, 1],
    "threads": 256,
    "blocks": 128,
    "intensity": 21
  },
  "opencl": {
    "enabled": false
  }
}
```

### Performance Tips

**NVIDIA GPUs:**
- RTX 3060 Ti: ~20-22 MH/s (90-120W)
- RTX 3070: ~23-25 MH/s (120-140W)
- RTX 3080: ~40-42 MH/s (220-250W)
- RTX 4090: ~60-65 MH/s (300-350W)

**AMD GPUs:**
- RX 6600 XT: ~14-16 MH/s (80-100W)
- RX 6800: ~26-28 MH/s (160-180W)
- RX 6900 XT: ~30-32 MH/s (200-230W)

**Optimization:**
- Underclock core slightly (-100 to -200 MHz)
- Increase memory clock (+500 to +1000 MHz)
- Reduce power limit (70-80%)
- Ensure adequate cooling

## ETChash

### Overview

ETChash (Etchash) is Ethereum Classic's mining algorithm, a variant of Ethash designed to be ASIC-resistant.

### Characteristics

- **Type:** GPU-optimized PoW (DAG-based)
- **Memory Requirement:** 5GB+ VRAM (increasing with DAG)
- **Optimal Hardware:** High-memory GPUs
- **ASIC Resistance:** Medium

### DAG Size

The DAG (Directed Acyclic Graph) increases over time:

| Date | DAG Size | Min VRAM |
|------|----------|----------|
| 2025 | ~5.0 GB | 6 GB |
| 2026 | ~5.3 GB | 6 GB |
| 2027 | ~5.6 GB | 8 GB |
| 2028 | ~5.9 GB | 8 GB |

### Configuration

```json
{
  "algo": "etchash",
  "cuda": {
    "enabled": true,
    "devices": [0],
    "threads": 256,
    "blocks": 128
  },
  "opencl": {
    "enabled": false
  }
}
```

### Performance Tips

**NVIDIA GPUs:**
- RTX 3060 Ti: ~60-62 MH/s (120-140W)
- RTX 3070: ~62-64 MH/s (120-140W)
- RTX 3080: ~100-105 MH/s (220-250W)
- RTX 3090: ~120-125 MH/s (280-320W)

**AMD GPUs:**
- RX 6600 XT: ~30-32 MH/s (60-75W)
- RX 6800: ~62-64 MH/s (120-140W)
- RX 6900 XT: ~64-66 MH/s (140-160W)

**Optimization:**
- Core clock: Moderate (not critical)
- Memory clock: High (critical for ETChash)
- Power limit: 60-75%
- Memory timings: Tight (if supported)

## ProgPowZ

### Overview

ProgPowZ is Zano's implementation of ProgPoW (Programmatic Proof-of-Work), designed to be ASIC-resistant through random program generation.

### Characteristics

- **Type:** GPU-optimized PoW
- **Memory Requirement:** 2-3GB VRAM
- **Optimal Hardware:** NVIDIA/AMD GPUs
- **ASIC Resistance:** High

### Configuration

```json
{
  "algo": "progpowz",
  "cuda": {
    "enabled": true,
    "devices": [0, 1],
    "threads": 256,
    "blocks": 64
  }
}
```

### Performance Tips

**NVIDIA GPUs:**
- RTX 3070: ~35-37 MH/s
- RTX 3080: ~55-60 MH/s
- RTX 4070: ~40-45 MH/s

**AMD GPUs:**
- RX 6700 XT: ~28-30 MH/s
- RX 6800: ~35-38 MH/s
- RX 6900 XT: ~40-42 MH/s

## Blake3

### Overview

Blake3 is a cryptographic hash function that can be mined on both CPU and GPU. It's used by Decred and other cryptocurrencies.

### Characteristics

- **Type:** Hybrid CPU/GPU PoW
- **Memory Requirement:** Low
- **Optimal Hardware:** Multi-core CPUs or GPUs
- **ASIC Resistance:** Low (ASICs available)

### Configuration

**CPU Mining:**
```json
{
  "algo": "blake3",
  "threads": 8,
  "cpuPriority": 3
}
```

**GPU Mining:**
```json
{
  "algo": "blake3",
  "cuda": {
    "enabled": true,
    "devices": [0]
  }
}
```

### Performance Tips

**CPU:**
- Modern CPUs with AVX2/AVX512: ~500-1500 MH/s per core
- Optimal for high core count processors

**GPU:**
- NVIDIA RTX 3080: ~15-20 GH/s
- AMD RX 6900 XT: ~12-15 GH/s

## CryptoNight

### Overview

CryptoNight is a memory-hard proof-of-work algorithm, formerly used by Monero before RandomX.

### Characteristics

- **Type:** Memory-hard PoW
- **Memory Requirement:** 2MB per thread
- **Optimal Hardware:** CPUs with AES-NI
- **ASIC Resistance:** Medium

### Variants

| Variant | Coin | Notes |
|---------|------|-------|
| cn/r | Monero (legacy) | CryptoNight R |
| cn/0 | Bytecoin | Original CryptoNight |
| cn/1 | MoneroV7 | CryptoNight v7 |
| cn/2 | MoneroV8 | CryptoNight v8 |
| cn/half | Masari | Half mode |
| cn/fast | Electroneum | Fast mode |

### Configuration

```json
{
  "algo": "cn/r",
  "threads": 4,
  "cpuPriority": 3,
  "aesNi": true
}
```

## Algorithm Selection

### By Hardware

**Strong CPU, No GPU:**
- Primary: RandomX (rx/0)
- Alternative: Blake3, CryptoNight

**Strong GPU(s), Weak CPU:**
- Primary: ETChash, KawPow, ProgPowZ
- Alternative: Blake3 (GPU)

**Both Strong CPU and GPU:**
- Dual mining: RandomX (CPU) + ETChash/KawPow (GPU)
- Maximum profitability

**Limited Hardware:**
- Start with: RandomX (low power CPU)
- Or: Blake3 (efficient on both)

### By Profitability

Check current profitability at:
- [WhatToMine](https://whattomine.com)
- [CoinWarz](https://www.coinwarz.com)
- [MiningPoolStats](https://miningpoolstats.stream)

Factors affecting profitability:
- Coin price
- Network difficulty
- Block rewards
- Pool fees
- Hardware efficiency

## Power Consumption

### Hashrate per Watt

**RandomX (CPU):**
- Typical: 20-40 H/s per Watt
- Efficient CPUs: 40-50 H/s per Watt

**KawPow (GPU):**
- NVIDIA RTX: 80-150 KH/s per Watt
- AMD RX: 100-180 KH/s per Watt

**ETChash (GPU):**
- NVIDIA RTX: 400-500 KH/s per Watt
- AMD RX: 400-600 KH/s per Watt

### Optimization for Efficiency

1. Reduce power limit (70-80% of max)
2. Find optimal core/memory clocks
3. Improve cooling (lower temps = better efficiency)
4. Use efficient PSU (80+ Gold or better)

## Dual Mining

### Compatible Combinations

**Recommended:**
- RandomX (CPU) + ETChash (GPU)
- RandomX (CPU) + KawPow (GPU)
- RandomX (CPU) + ProgPowZ (GPU)

**Configuration:**
```json
{
  "cpu": {
    "enabled": true,
    "algo": "rx/0",
    "threads": 6,
    "pool": "stratum+tcp://pool.supportxmr.com:3333",
    "wallet": "YOUR_XMR_WALLET"
  },
  "gpu": {
    "enabled": true,
    "algo": "etchash",
    "pool": "stratum+tcp://etc.woolypooly.com:3333",
    "wallet": "YOUR_ETC_WALLET",
    "cuda": {
      "devices": [0, 1]
    }
  }
}
```

## Benchmarking

Test your hardware performance:

```bash
# CPU benchmark
miner-ctrl benchmark --algo rx/0 --threads 8

# GPU benchmark
miner-ctrl benchmark --algo etchash --cuda --device 0

# Compare algorithms
miner-ctrl benchmark --all
```

## Troubleshooting

### Low CPU Hashrate

- Enable huge pages
- Increase CPU priority
- Reduce thread count
- Check for thermal throttling
- Close background applications

### Low GPU Hashrate

- Update GPU drivers
- Increase power limit
- Optimize memory clocks
- Check for thermal throttling
- Verify VRAM is sufficient for DAG

### High Rejected Shares

- Check network connection
- Reduce difficulty (use lower port)
- Verify algorithm matches pool
- Update mining software

## Resources

- [Pool Integration Guide](pools.md)
- [Quick Start Guide](../getting-started/quick-start.md)
- [Hardware Guides](https://www.reddit.com/r/MoneroMining)
- [Profitability Calculators](https://whattomine.com)

## Next Steps

- Read the [Pool Integration Guide](pools.md)
- Try the [Quick Start Guide](../getting-started/quick-start.md)
- Explore the [API Documentation](../api/endpoints.md)
