# Changelog

All notable changes to this project will be documented in this file.

- [1.1.0 (2024-07-14)](#110-2024-07-14)
- [1.0.2 (2023-05-28)](#102-2023-05-28)
- [1.0.1 (2023-05-27)](#101-2023-05-27)
- [1.0.0 (2023-05-27)](#100-2023-05-27)

---

<a name="1.1.0"></a>
## [1.1.0](https://github.com/aisbergg/go-bruh/compare/v1.0.2...v1.1.0) (2024-07-14)

### Bug Fixes

- properly identify and exclude globally defined errors again

### Features

- add `RangeContext` to ContextableError and improve general performance by using a slice as a backing for the context data
- add ProgramCounter2 that indicates the actual instruction being executed and not the one beforehand


<a name="1.0.2"></a>
## [1.0.2](https://github.com/aisbergg/go-bruh/compare/v1.0.1...v1.0.2) (2023-05-28)

### Performance Improvements

- improve performance of FormatWithCombinedTrace, FormatPythonTraceback and Stack.String()


<a name="1.0.1"></a>
## [1.0.1](https://github.com/aisbergg/go-bruh/compare/v1.0.0...v1.0.1) (2023-05-27)

### Documentation

- improve examples


<a name="1.0.0"></a>
## [1.0.0]() (2023-05-27)

Initial Release
