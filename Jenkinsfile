@Library('dst-shared@release/shasta-1.3') _
rpmBuild(
    product: "shasta-standard,shasta-premium",
    target_node: "ncn",
    channel: "casmcms-builds",
    fanout_params: ["sle15sp1"],
    recv_triggers: ["cme-premium-cf"]
)
