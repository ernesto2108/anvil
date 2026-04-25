# Patrones de Blockchain y Crypto

## Solana / Anchor

```rust
use anchor_lang::prelude::*;

declare_id!("YourProgramId11111111111111111111111111111");

#[program]
pub mod my_program {
    use super::*;

    pub fn initialize(ctx: Context<Initialize>, data: u64) -> Result<()> {
        let account = &mut ctx.accounts.my_account;
        account.data = data;
        account.authority = ctx.accounts.authority.key();
        Ok(())
    }

    pub fn update(ctx: Context<Update>, new_data: u64) -> Result<()> {
        let account = &mut ctx.accounts.my_account;
        require!(
            account.authority == ctx.accounts.authority.key(),
            MyError::Unauthorized
        );
        account.data = new_data;
        Ok(())
    }
}

#[derive(Accounts)]
pub struct Initialize<'info> {
    #[account(init, payer = authority, space = 8 + MyAccount::INIT_SPACE)]
    pub my_account: Account<'info, MyAccount>,
    #[account(mut)]
    pub authority: Signer<'info>,
    pub system_program: Program<'info, System>,
}

#[account]
#[derive(InitSpace)]
pub struct MyAccount {
    pub authority: Pubkey,
    pub data: u64,
}

#[error_code]
pub enum MyError {
    #[msg("Unauthorized")]
    Unauthorized,
}
```

## Alloy (Ethereum, reemplaza ethers-rs)

```rust
use alloy::{providers::ProviderBuilder, primitives::{Address, U256}, sol};

sol! {
    #[sol(rpc)]
    contract ERC20 {
        function balanceOf(address owner) external view returns (uint256);
        function transfer(address to, uint256 amount) external returns (bool);
        event Transfer(address indexed from, address indexed to, uint256 value);
    }
}

async fn check_balance(rpc: &str, token: Address, wallet: Address) -> anyhow::Result<U256> {
    let provider = ProviderBuilder::new().on_http(rpc.parse()?);
    let contract = ERC20::new(token, &provider);
    let balance = contract.balanceOf(wallet).call().await?;
    Ok(balance._0)
}
```

## Primitivos Criptográficos (RustCrypto)

```rust
use sha2::{Sha256, Digest};
use ed25519_dalek::{SigningKey, Signer, Verifier};
use subtle::ConstantTimeEq;

fn hash(data: &[u8]) -> [u8; 32] {
    let mut hasher = Sha256::new();
    hasher.update(data);
    hasher.finalize().into()
}

// WRONG — timing side-channel
fn verify(expected: &[u8], computed: &[u8]) -> bool {
    expected == computed  // short-circuits
}

// RIGHT — constant-time comparison
fn verify(expected: &[u8; 32], computed: &[u8; 32]) -> bool {
    expected.ct_eq(computed).into()
}
```

## no_std para Embedded/WASM

```rust
#![cfg_attr(not(feature = "std"), no_std)]

#[cfg(not(feature = "std"))]
extern crate alloc;
#[cfg(not(feature = "std"))]
use alloc::{string::String, vec::Vec};

#[derive(Debug)]
pub enum Error {
    InvalidInput,
    CryptoError,
}

impl core::fmt::Display for Error {
    fn fmt(&self, f: &mut core::fmt::Formatter<'_>) -> core::fmt::Result {
        match self {
            Error::InvalidInput => write!(f, "invalid input"),
            Error::CryptoError => write!(f, "crypto error"),
        }
    }
}
```

## Crates Clave

- `alloy` 0.9.x — Ethereum (reemplaza ethers-rs)
- `anchor-lang` 0.30.x — programas Solana
- `solana-sdk` 2.x — SDK base de Solana
- `sha2`/`sha3` 0.10.x — hashing (RustCrypto)
- `ed25519-dalek` 2.x — firmas Ed25519
- `k256` 0.13.x — secp256k1 (curva Ethereum)
- `aes-gcm` 0.10.x — cifrado AES-GCM
- `subtle` 2.x — operaciones en tiempo constante
- `rand` 0.8.x + `rand_chacha` — CSPRNG
- `secrecy` 0.10.x — zeroize al soltar
