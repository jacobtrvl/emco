```
Copyright (c) 2022 Intel Corporation
```

# EMCO Port Numbers

Ports fall in one of 3 categories:

- **Service** (RESTful HTTP API)
- **Control** (gRPC)
- **Status** (gRPC)

Every EMCO **Service** port number is configurable. gRPC-based port numbers may be configurable in some cases, but they are mostly defined in the source code (at the time of writing).
For any port that isn't configured (or configurable), a default will be used.

Here is the list of all EMCO services, controllers and subcontrollers where port numbers are used, and respective defaults for the internal (local) ports used by each:

    CLM_SERVICE_PORT=9061
    DCM_SERVICE_PORT=9077
    DCM_STATUS_PORT=9078
    DTC_CONTROL_PORT=9048
    DTC_SERVICE_PORT=9018
    GAC_CONTROL_PORT=9033
    GAC_SERVICE_PORT=9020
    HPAAC_CONTROL_PORT=9042
    HPAAC_SERVICE_PORT=9041
    HPAPLC_CONTROL_PORT=9099
    HPAPLC_SERVICE_PORT=9091
    ITS_CONTROL_PORT=9040
    ITS_SERVICE_PORT=9037
    NCM_SERVICE_PORT=9081
    NCM_STATUS_PORT=9082
    NPS_CONTROL_PORT=9038
    NPS_SERVICE_PORT=9036
    ORCH_SERVICE_PORT=9015
    ORCH_STATUS_PORT=9016
    OVN_CONTROL_PORT=9032
    OVN_SERVICE_PORT=9051
    RSYNC_SERVICE_PORT=9023
    RSYNC_CONTROL_PORT=9031
    SDS_CONTROL_PORT=9039
    SDS_SERVICE_PORT=9035
    SFC_CONTROL_PORT=9056
    SFC_SERVICE_PORT=9055
    SFCC_CONTROL_PORT=9058
    SFCC_SERVICE_PORT=9057
    SWC_CONTROL_PORT=9088
    SWC_SERVICE_PORT=9087

And here is the equivalent list, but for the default node ports used by each, when EMCO is deployed using official installation guides:

    CLM_SERVICE_PORT=30461
    DCM_SERVICE_PORT=30477
    DCM_STATUS_PORT=30478
    DTC_CONTROL_PORT=30448
    DTC_SERVICE_PORT=30418
    GAC_CONTROL_PORT=30433
    GAC_SERVICE_PORT=30420
    HPAAC_CONTROL_PORT=30442
    HPAAC_SERVICE_PORT=30441
    HPAPLC_CONTROL_PORT=30499
    HPAPLC_SERVICE_PORT=30491
    ITS_CONTROL_PORT=30440
    ITS_SERVICE_PORT=30437
    NCM_SERVICE_PORT=30481
    NCM_STATUS_PORT=30482
    NPS_CONTROL_PORT=30438
    NPS_SERVICE_PORT=30436
    ORCH_SERVICE_PORT=30415
    ORCH_STATUS_PORT=30416
    OVN_CONTROL_PORT=30432
    OVN_SERVICE_PORT=30451
    RSYNC_SERVICE_PORT=30423
    RSYNC_CONTROL_PORT=30431
    SDS_CONTROL_PORT=30439
    SDS_SERVICE_PORT=30435
    SFC_CONTROL_PORT=30456
    SFC_SERVICE_PORT=30455
    SFCC_CONTROL_PORT=30458
    SFCC_SERVICE_PORT=30457
    SWC_CONTROL_PORT=30488
    SWC_SERVICE_PORT=30487

The lists above have no intention of claiming that each of those default ports numbers can be overriden by configuration. They are simply communicating every single port that will be used by EMCO, if nothing is modified/configured in that regard.

Both of the lists above are presented in the format accepted by the config files used within the `examples` directory (as of this writing, only `examples/single-cluster` has been migrated to make use of config files).

All port numbers follow the format of:

- `90` followed by a 2-digit number unique to the particular service (like DCM) and category (such as "status"), for internal ports (DCM "status" thus `9078`)
- `304` followed by a 2-digit number unique to the particular service (like DCM) and category (such as "status"), for node ports (DCM "status" thus `30478`)
