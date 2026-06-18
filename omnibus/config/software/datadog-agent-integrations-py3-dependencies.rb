name 'datadog-agent-integrations-py3-dependencies'

dependency 'pip3'
dependency 'setuptools3'

if linux_target?

  build do
    flavor_flag = fips_mode? ? "--//packages/agent:flavor=fips" : ""

    # [sts] STAC-24773 C5: unixodbc + freetds + msodbcsql18 migrated to Bazel
    # (replacing omnibus freetds.rb / msodbcsql18.rb / unixodbc.rb). Mirrors
    # origin/base-7.78.2 with --downloader_config=/dev/null for STS runners.
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @unixodbc//:install --destdir='#{install_dir}'"
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
      " #{install_dir}/embedded/lib/libodbc.so" \
      " #{install_dir}/embedded/lib/libodbccr.so" \
      " #{install_dir}/embedded/lib/libodbcinst.so"

    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- @freetds//:install --destdir='#{install_dir}'"
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded'" \
      " #{install_dir}/embedded/lib/libtdsodbc.so"

    unless heroku_target?
      lib_files = [
        'krb5/plugins/tls/k5tls.so',
        'krb5/plugins/kdb/db2.so',
        'krb5/plugins/preauth/test.so',
        'krb5/plugins/preauth/spake.so',
        'krb5/plugins/preauth/pkinit.so',
        'krb5/plugins/preauth/otp.so',
        'libkadm5clnt_mit.so',
        'libkrad.so',
        'libverto.so',
        'libk5crypto.so',
        'libcom_err.so',
        'libkadm5srv.so',
        'libkrb5support.so',
        'libgssrpc.so',
        'libkrb5.so',
        'libkadm5srv_mit.so',
        'libkdb5.so',
        'libgssapi_krb5.so',
        'libkadm5clnt.so',
      ]
      bin_files = [
        'kinit',
      ]

      command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //deps/msodbcsql18:install --destdir='#{install_dir}'"
      command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //bazel/rules:replace_prefix --prefix '#{install_dir}/embedded' " \
        + lib_files.map { |l| "#{install_dir}/embedded/lib/#{l}" }.join(' ') \
        + " " \
        + bin_files.map { |bin| "#{install_dir}/embedded/bin/#{bin}" }.join(' ') \
        + " '#{install_dir}/embedded/msodbcsql/lib64/libmsodbcsql-18.3.so.3.1'"
    end

    # [sts] STAC-24773 C6: gstatus + nfsiostat via Bazel (replaces omnibus .rb files).
    # deps/{gstatus,nfsiostat}/BUILD.bazel hardcode /opt/datadog-agent/ in shebangs;
    # re-stamp to #{install_dir} post-install (same pattern as old gstatus.rb).
    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //deps/gstatus:install --destdir='#{install_dir}'"
    command "sed -i '1s|.*|#!#{install_dir}/embedded/bin/python|' #{install_dir}/embedded/sbin/gstatus"

    command_on_repo_root "bazelisk run #{flavor_flag} --downloader_config=/dev/null -- //deps/nfsiostat:install --destdir='#{install_dir}'"
    command "sed -i '1s|.*|#!#{install_dir}/embedded/bin/python|' #{install_dir}/embedded/sbin/nfsiostat"
  end
end
