# Contributing to _bruh_

Thank you for considering contributing to _bruh_! We appreciate your interest and support in making this library better. This document provides some guidelines on how you can contribute to the project.

<details>
  <summary>Table of Contents</summary>

- [How Can I Contribute?](#how-can-i-contribute)
- [Reporting Issues](#reporting-issues)
- [Submitting Pull Requests](#submitting-pull-requests)
- [Coding Guidelines](#coding-guidelines)
- [Documentation](#documentation)
- [Code of Conduct](#code-of-conduct)

</details>



## How Can I Contribute?

There are several ways you can contribute to _bruh_:

- Reporting issues: If you encounter any bugs or have suggestions for improvements, please let us know by creating an issue in the GitHub repository.
- Submitting pull requests: If you'd like to contribute code, you can fork the repository, make your changes, and submit a pull request for review.
- Improving documentation: Documentation is vital for any library, and you can help improve it by fixing typos, clarifying explanations, or adding new examples.
- Providing feedback: If you have ideas or suggestions for new features, enhancements, or changes, you can share them by creating a GitHub issue.

## Reporting Issues

If you encounter any issues with _bruh_, please report them using the following steps:

1. Ensure the issue hasn't already been reported by searching through the existing issues on GitHub.
2. If you don't find an existing issue, create a new one [here](https://github.com/aisbergg/go-bruh/issues/new/choose). Choose an issue template that seems most appropriate for your case. Provide a descriptive title and a clear explanation of the problem.
3. Be patient and wait for feedback. We try to tend to your request in a reasonable time, but we cannot make guarantees. Our personal life takes priority and therefore don't get mad, if it might take a while, before you hear back from us.

## Submitting Pull Requests

We welcome contributions in the form of pull requests. To submit a pull request:

1. Fork the repository.
2. Create a new branch from the `main` branch for your changes. (`git checkout -b feat/amazing-feature main`)
2. Install the `pre-commit` hooks and other development tools.

    using make:
    ```sh
    make dev-setup
    ```

    manually:
    ```sh
    # install requirements
    pip3 install -u pre-commit
    bash tools/install.sh

    # enable hooks for this repo
    pre-commit install
    ```
3. Make your changes, ensuring that your code follows the coding guidelines (discussed later in this document).
4. Write tests to cover your changes and ensure they pass successfully.
5. Commit your changes with clear and descriptive commit messages. Please use the [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/) style for the messages. (`git commit -m 'feat: add amazing feature'`)
6. Push your branch to your forked repository. (`git push origin feat/amazing-feature`)
7. Open a pull request on the main repository and provide a detailed description of your changes.
8. Be responsive to any feedback or review comments during the review process.

## Coding Guidelines

When contributing code to _bruh_, please follow these coding guidelines:

- Use idiomatic Go code that adheres to the official Go [Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments).
- Format your code using `gofmt` or any compatible code formatter.
- Write clear, concise, and self-explanatory code.
- Add meaningful comments to complex or non-obvious parts of your code.
- Ensure your code passes all existing tests, and write new tests for your changes.

## Documentation

Documentation is crucial for helping users understand and effectively use _bruh_. If you'd like to contribute to the documentation:

- Ensure your changes align with the library's purpose and scope.
- Use clear and concise language to explain concepts, features, and examples.
- Keep the documentation up to date with the latest changes in the library.
- Proofread your changes to ensure accuracy and correct any typos or grammatical errors.

## Code of Conduct

To maintain a positive and inclusive community, we follow the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org). We expect all contributors and participants in the _bruh_ community to adhere to the Code of Conduct. In summary:

- Be respectful and considerate towards others, regardless of their background, experience, or opinions.
- Refrain from any form of harassment, discrimination, or offensive behavior.
- Foster a welcoming and inclusive environment for all individuals.
- Communicate in a constructive and professional manner.
- Respect the project's guidelines, processes, and decisions.

You can find a copy of the Code of Conduct in the [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) document.
