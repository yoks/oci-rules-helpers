"""Generate a image patch for flux oci image
"""

def setup_oci_image(**kwargs):
    _setup_oci_image(
        **kwargs
    )

def _setup_oci_image_impl(ctx):
    push_info = ctx.attr.image[DefaultInfo]

    inputs = []
    image_manifest = ctx.actions.declare_file("{}/manifest.json".format(
        ctx.label.name,
    ))
    inputs.append(image_manifest)
    ctx.actions.write(
        output = image_manifest,
        content = json.encode_indent(
            [manifest.path for manifest in push_info.default_runfiles.files.to_list()]
        ),
    )
    image_runfiles = push_info.default_runfiles.files.to_list()
    inputs.extend(image_runfiles)

    args = ctx.actions.args()

    args.add("-template", ctx.file.template)
    args.add("-out", ctx.outputs.out)
    args.add("-manifest", image_manifest)

    ctx.actions.run(
        executable = ctx.executable._writer,
        outputs = [ctx.outputs.out],
        inputs = depset(
            inputs + [ctx.file.template],
        ),
        arguments = [args],
        progress_message = "Creating image file for {}".format(
            ctx.label,
        ),
    )

    return [
        DefaultInfo(
            files = depset([ctx.outputs.out]),
            runfiles = ctx.runfiles([ctx.outputs.out]),
        ),
    ]

_setup_oci_image = rule(
    implementation = _setup_oci_image_impl,
    attrs = {
        "image": attr.label(mandatory = True, allow_single_file = True),
        "template": attr.label(
            allow_single_file = True,
        ),
        "out": attr.output(mandatory = True),
        "_writer": attr.label(
            cfg = "exec",
            executable = True,
            default = Label("//bazel/rules/image:write_image_file"),
        ),
    },
)
