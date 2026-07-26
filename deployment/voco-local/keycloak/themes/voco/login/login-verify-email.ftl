<#import "template.ftl" as layout>
<@layout.registrationLayout displayInfo=true; section>
    <#if section = "header">
        ${msg("emailVerifyTitle")}
    <#elseif section = "form">
        <div class="voco-info-panel">
            <p class="instruction voco-info-text">
                <#if verifyEmail??>
                    ${msg("emailVerifyInstruction1",verifyEmail)}
                <#else>
                    ${msg("emailVerifyInstruction4",user.email)}
                </#if>
            </p>

            <#if isAppInitiatedAction??>
                <form id="kc-verify-email-form" class="${properties.kcFormClass!}" action="${url.loginAction}" method="post">
                    <div class="${properties.kcFormGroupClass!}">
                        <div id="kc-form-buttons" class="voco-otp-actions">
                            <#if verifyEmail??>
                                <input class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!}" type="submit" value="${msg("emailVerifyResend")}" />
                            <#else>
                                <input class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!}" type="submit" value="${msg("emailVerifySend")}" />
                            </#if>
                            <button class="${properties.kcButtonClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!} voco-btn-secondary" type="submit" name="cancel-aia" value="true" formnovalidate>${msg("doCancel")}</button>
                        </div>
                    </div>
                </form>
            <#else>
                <a class="pf-v5-c-button pf-m-primary pf-m-block voco-cta" href="${url.loginAction}">${msg("emailVerifyCheckAgain")}</a>
                <p class="instruction voco-info-text voco-info-secondary">
                    ${msg("emailVerifyInstruction2")}
                    <a href="${url.loginAction}">${msg("doClickHere")}</a> ${msg("emailVerifyInstruction3")}
                </p>
            </#if>
        </div>
    </#if>
</@layout.registrationLayout>
