#[cfg(target_os = "windows")]
fn main() {
    use core::panic;
    use std::path::Path;

    let vcpkg_dir = format!("{}/vcpkg_installed", std::env::current_dir().unwrap().to_string_lossy());
    let bin_dir = format!("{}/x64-windows/bin", vcpkg_dir);
    let lib_dir = format!("{}/x64-windows/lib", vcpkg_dir);
    if !Path::new(vcpkg_dir.as_str()).exists() {
        panic!("Please run vcpkg install before building");
    }

    println!("cargo:rustc-link-lib=dylib=SDL2");
    println!("cargo:rustc-link-search=native={}", bin_dir);
    println!("cargo:rustc-link-search=native={}", lib_dir);
}

#[cfg(not(target_os = "windows"))]
fn main() {}
