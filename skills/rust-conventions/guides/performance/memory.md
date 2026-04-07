# Memory & Performance Patterns

## Zero-Copy Parsing

```rust
// WRONG — allocating during parse
fn parse_header(input: &[u8]) -> String {
    String::from_utf8(input[..32].to_vec()).unwrap()
}

// RIGHT — borrow from input
fn parse_header(input: &[u8]) -> &str {
    std::str::from_utf8(&input[..32]).unwrap()
}

// RIGHT — bytes crate for zero-copy buffers
use bytes::{Bytes, BytesMut, Buf};

fn parse_frame(buf: &mut BytesMut) -> Option<Bytes> {
    if buf.len() < 4 { return None; }
    let len = u32::from_be_bytes(buf[..4].try_into().unwrap()) as usize;
    if buf.len() < 4 + len { return None; }
    buf.advance(4);
    Some(buf.split_to(len).freeze())  // zero-copy
}
```

## Cow (Conditional Ownership)

```rust
use std::borrow::Cow;

// WRONG — always allocates
fn normalize(input: &str) -> String {
    if input.contains(' ') {
        input.replace(' ', "_")
    } else {
        input.to_string()  // unnecessary
    }
}

// RIGHT — borrows when possible
fn normalize(input: &str) -> Cow<'_, str> {
    if input.contains(' ') {
        Cow::Owned(input.replace(' ', "_"))
    } else {
        Cow::Borrowed(input)
    }
}
```

## Arena Allocation

```rust
use bumpalo::Bump;

fn process_batch(records: &[RawRecord]) -> Vec<Processed> {
    let arena = Bump::new();
    let mut results = Vec::new();
    for record in records {
        let temp = arena.alloc_str(&record.text);
        let parsed = parse_in_arena(&arena, temp);
        results.push(parsed.to_owned());
    }
    results
    // arena drops here — all temp allocations freed at once
}
```

## SmallVec

```rust
use smallvec::SmallVec;

// Stack-allocated for small sizes, heap for large
type Tags = SmallVec<[Tag; 4]>;

fn get_tags(item: &Item) -> Tags {
    let mut tags = SmallVec::new();
    tags.push(item.primary_tag());
    tags  // stays on stack if <= 4 tags
}
```

## Unsafe Guidelines

```rust
// WRONG — large unsafe block, no comment
unsafe {
    let ptr = alloc(layout);
    ptr::write(ptr as *mut Header, header);
    let slice = slice::from_raw_parts_mut(ptr.add(HEADER_SIZE), len);
    slice.copy_from_slice(data);
}

// RIGHT — minimize scope, document each invariant
let ptr = unsafe { alloc(layout) };
if ptr.is_null() { return Err(AllocError); }

// SAFETY: ptr is non-null, aligned for Header, allocation is large enough
unsafe { ptr::write(ptr as *mut Header, header) };

// SAFETY: ptr + HEADER_SIZE is within allocation, len bytes available
let slice = unsafe { slice::from_raw_parts_mut(ptr.add(HEADER_SIZE), len) };
slice.copy_from_slice(data);  // safe — outside unsafe
```

## Key Crates

- `bytes` 1.x — zero-copy byte buffers
- `bumpalo` 3.x — arena allocator
- `smallvec` 1.x — stack-allocated small vectors
- `arrayvec` 0.7.x — fixed-capacity stack vectors
- `memmap2` 0.9.x — memory-mapped files
- `zerocopy` 0.8.x — zero-copy structured data
- `criterion` 0.5.x — benchmarking
