//! DNET Wasm Edge Node - Constant-Time Sensor Payload Parser
//!
//! SECURITY MODEL: Timing side-channel resistant parsing for adversarial environments.
//! All code paths execute in constant time regardless of input validity.

#![no_std]

/// A parsed sensor reading from a DNET Wasm edge node.
#[derive(Clone, Copy, Debug, PartialEq, Eq)]
pub struct SensorReading {
    pub value: u32,
    pub state: u8,
    pub is_valid: bool,
}

/// Constant-time byte equality-to-zero check.
/// Returns 1 if x == 0, else 0.
///
/// # Mathematical Basis
///
/// In two's complement arithmetic, for any non-zero byte x:
///   - If 0 < x ≤ 127: wrapping_neg(x) = 256 - x, which has MSB set (value ≥ 129)
///   - If 128 ≤ x ≤ 255: x already has MSB set
///
/// Therefore: (x | x.wrapping_neg()) has MSB set ⟺ x ≠ 0
///
/// Shifting right by 7 extracts the MSB: 1 for non-zero, 0 for zero.
/// We invert via wrapping_sub to get: 1 for zero, 0 for non-zero.
///
/// # Timing Properties
/// - Always executes: 1 OR, 1 NEG, 1 SHIFT, 1 SUB
/// - No data-dependent paths
#[inline(always)]
fn ct_is_zero_u8(x: u8) -> u8 {
    (1u8).wrapping_sub((x | x.wrapping_neg()) >> 7)
}

/// Constant-time 16-bit equality-to-zero check.
/// Returns 1 if x == 0, else 0.
///
/// # Strategy
/// Reduce to single byte by ORing high and low bytes together.
/// If either byte is non-zero, the result is non-zero.
/// Then apply the byte-wise constant-time check.
///
/// # Timing Properties
/// - Always executes: 2 SHIFTs, 1 OR (16-bit), 1 cast, 1 call to ct_is_zero_u8
#[inline(always)]
fn ct_is_zero_u16(x: u16) -> u8 {
    let reduced = (x >> 8) as u8 | (x as u8);
    ct_is_zero_u8(reduced)
}

/// Parses an 8-byte DNET sensor payload in constant time.
///
/// # Payload Structure
/// ```text
/// +--------+--------+-----------------+-----------------+
/// | Byte 0 | Byte 1 |    Bytes 2-5    |    Bytes 6-7    |
/// +--------+--------+-----------------+-----------------+
/// | Magic  | State  | Value (BE u32)  | CSum (BE u16)   |
/// | 0xAA   | 0/1    |                 | sum(bytes 0..5) |
/// +--------+--------+-----------------+-----------------+
/// ```
///
/// # Security Guarantees
/// - **No branching**: Zero conditional control flow
/// - **No short-circuits**: All validation checks always execute
/// - **Constant latency**: Valid and invalid payloads take identical time
///
/// # Validation Rules
/// 1. Magic byte must equal 0xAA
/// 2. State must be 0x00 or 0x01
/// 3. Checksum must equal sum of bytes 0-5
#[inline(never)]
#[no_mangle]
pub fn parse_sensor_payload(payload: [u8; 8]) -> SensorReading {
    // ============================================================
    // PHASE 1: Field Extraction (pure bitwise arithmetic)
    // ============================================================
    
    // Extract sensor value: big-endian u32 from bytes [2,3,4,5]
    let value: u32 = (payload[2] as u32) << 24
                   | (payload[3] as u32) << 16
                   | (payload[4] as u32) << 8
                   | (payload[5] as u32);
    
    // Extract state: raw byte from position 1
    let state: u8 = payload[1];
    
    // Extract provided checksum: big-endian u16 from bytes [6,7]
    let checksum: u16 = (payload[6] as u16) << 8 | (payload[7] as u16);
    
    // Compute expected checksum: wrapping sum of bytes 0 through 5
    // Maximum possible: 6 × 255 = 1,530 < 65,535 (fits in u16)
    // Using wrapping_add to prevent compiler overflow checks (branch-free)
    let expected: u16 = (payload[0] as u16)
                      .wrapping_add(payload[1] as u16)
                      .wrapping_add(payload[2] as u16)
                      .wrapping_add(payload[3] as u16)
                      .wrapping_add(payload[4] as u16)
                      .wrapping_add(payload[5] as u16);
    
    // ============================================================
    // PHASE 2: Constant-Time Validation
    // ============================================================
    // Each check produces 1 (pass) or 0 (fail).
    // We combine with bitwise operators (no short-circuit).
    
    // --- Check 1: Magic byte validation ---
    // XOR with expected value: 0 if match, non-zero otherwise
    // ct_is_zero_u8 converts to 1/0 flag
    let magic_xor = payload[0] ^ 0xAA;
    let magic_valid: u8 = ct_is_zero_u8(magic_xor);
    
    // --- Check 2: State field validation ---
    // State must be in {0x00, 0x01}
    // Check each possibility independently, OR the results
    let state_eq_0: u8 = ct_is_zero_u8(state ^ 0x00);
    let state_eq_1: u8 = ct_is_zero_u8(state ^ 0x01);
    let state_valid: u8 = state_eq_0 | state_eq_1;
    
    // --- Check 3: Checksum validation ---
    // XOR provided vs expected: 0 if match, non-zero otherwise
    let checksum_xor = checksum ^ expected;
    let checksum_valid: u8 = ct_is_zero_u16(checksum_xor);
    
    // --- Combine all checks ---
    // Bitwise AND: all three must be 1 for overall validity
    // Result is 0 (invalid) or 1 (valid)
    let all_valid: u8 = magic_valid & state_valid & checksum_valid;
    
    // --- Convert to boolean ---
    // Comparison generates setne instruction, not a branch
    // Alternatively: unsafe { core::mem::transmute::<u8, bool>(all_valid) }
    let is_valid: bool = all_valid != 0;
    
    SensorReading { value, state, is_valid }
}

// ============================================================
// VERIFICATION: Unit Tests (compile with --tests)
// ============================================================

#[cfg(test)]
mod tests {
    use super::*;
    
    fn build_payload(magic: u8, state: u8, value: u32, override_checksum: Option<u16>) -> [u8; 8] {
        let mut payload = [0u8; 8];
        payload[0] = magic;
        payload[1] = state;
        payload[2] = (value >> 24) as u8;
        payload[3] = (value >> 16) as u8;
        payload[4] = (value >> 8) as u8;
        payload[5] = value as u8;
        
        let checksum = override_checksum.unwrap_or_else(|| {
            (magic as u16 + state as u16 + payload[2] as u16 
             + payload[3] as u16 + payload[4] as u16 + payload[5] as u16)
        });
        
        payload[6] = (checksum >> 8) as u8;
        payload[7] = checksum as u8;
        payload
    }
    
    #[test]
    fn valid_payload_state_zero() {
        let payload = build_payload(0xAA, 0x00, 0x12345678, None);
        let reading = parse_sensor_payload(payload);
        
        assert_eq!(reading.value, 0x12345678);
        assert_eq!(reading.state, 0x00);
        assert!(reading.is_valid);
    }
    
    #[test]
    fn valid_payload_state_one() {
        let payload = build_payload(0xAA, 0x01, 0xDEADBEEF, None);
        let reading = parse_sensor_payload(payload);
        
        assert_eq!(reading.value, 0xDEADBEEF);
        assert_eq!(reading.state, 0x01);
        assert!(reading.is_valid);
    }
    
    #[test]
    fn invalid_magic_byte() {
        let payload = build_payload(0xBB, 0x01, 0x12345678, None);
        let reading = parse_sensor_payload(payload);
        
        assert_eq!(reading.value, 0x12345678);
        assert!(!reading.is_valid, "Should be invalid: wrong magic byte");
    }
    
    #[test]
    fn invalid_state_out_of_range() {
        let payload = build_payload(0xAA, 0x02, 0x12345678, None);
        let reading = parse_sensor_payload(payload);
        
        assert!(!reading.is_valid, "Should be invalid: state = 0x02");
    }
    
    #[test]
    fn invalid_state_high_bit() {
        let payload = build_payload(0xAA, 0xFF, 0x12345678, None);
        let reading = parse_sensor_payload(payload);
        
        assert!(!reading.is_valid, "Should be invalid: state = 0xFF");
    }
    
    #[test]
    fn invalid_checksum() {
        let payload = build_payload(0xAA, 0x01, 0x12345678, Some(0x0000));
        let reading = parse_sensor_payload(payload);
        
        assert!(!reading.is_valid, "Should be invalid: wrong checksum");
    }
    
    #[test]
    fn multiple_failures() {
        let mut payload = [0xBB, 0xFF, 0x12, 0x34, 0x56, 0x78, 0x00, 0x00];
        let reading = parse_sensor_payload(payload);
        
        assert!(!reading.is_valid, "Should be invalid: multiple fields wrong");
    }
    
    #[test]
    fn boundary_value_max() {
        // Maximum valid values: magic=0xAA, state=0x01, value=0xFFFFFFFF
        // Checksum = 0xAA + 0x01 + 0xFF + 0xFF + 0xFF + 0xFF = 0x4FC
        let payload = build_payload(0xAA, 0x01, 0xFFFFFFFF, None);
        let reading = parse_sensor_payload(payload);
        
        assert_eq!(reading.value, 0xFFFFFFFF);
        assert!(reading.is_valid);
    }
    
    #[test]
    fn boundary_value_zero() {
        let payload = build_payload(0xAA, 0x00, 0x00000000, None);
        let reading = parse_sensor_payload(payload);
        
        assert_eq!(reading.value, 0x00000000);
        assert!(reading.is_valid);
    }
}
