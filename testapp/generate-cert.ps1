$rootCert = New-SelfSignedCertificate -DnsName "gcpupdater" -Type CodeSigningCert -CertStoreLocation ".\cert"
[System.Security.SecureString]$rootcertPassword = ConvertTo-SecureString -String "thelongandwindingsecret" -Force -AsPlainText
[String]$rootCertPath = Join-Path -Path 'cert:\CurrentUser\My\' -ChildPath "$($rootcert.Thumbprint)"
[String]$rootCertPath = Join-Path -Path 'cert:\LocalMachine\My\' -ChildPath "$($rootcert.Thumbprint)"
Export-PfxCertificate -Cert $rootCertPath -FilePath '.\cert\testcert.pfx' -Password $rootcertPassword
Export-Certificate -Cert $rootCertPath -FilePath '.\cert\testcert.crt'