package(default_visibility = ["//visibility:public"])

licenses(["notice"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_binary",
    "go_library",
)

go_binary(
    name = "kubeadm",
    library = ":go_default_library",
    linkstamp = "github.com/sourcegraph/monorepo-test-1/kubernetes-4/pkg/version",
    tags = ["automanaged"],
)

go_library(
    name = "go_default_library",
    srcs = ["kubeadm.go"],
    tags = ["automanaged"],
    deps = ["//cmd/kubeadm/app:go_default_library"],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//cmd/kubeadm/app:all-srcs",
        "//cmd/kubeadm/test/cmd:all-srcs",
    ],
    tags = ["automanaged"],
)
