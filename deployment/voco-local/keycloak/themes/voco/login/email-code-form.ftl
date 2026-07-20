<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=!messagesPerField.existsError('emailCode'); section>
    <#if section="header">
        ${msg("emailOtpTitle")}
    <#elseif section="form">
        <form id="kc-otp-login-form" class="${properties.kcFormClass!}" action="${url.loginAction}" method="post">
            <#assign otpLength = (codeLength!6)>

            <p class="instruction voco-otp-hint">
                ${msg("emailOtpForm", otpLength)}<#if maskedEmail??>: <strong>${kcSanitize(maskedEmail)?no_esc}</strong></#if>
            </p>

            <div class="${properties.kcFormGroupClass!}">
                <div class="voco-otp-cells" data-otp-length="${otpLength}" data-otp-target="emailCode">
                    <#list 0..<otpLength as i>
                        <input
                            type="text"
                            inputmode="numeric"
                            pattern="[0-9]*"
                            maxlength="1"
                            class="voco-otp-cell"
                            autocomplete="<#if i == 0>one-time-code<#else>off</#if>"
                            aria-label="${msg("emailOtpDigit", (i + 1)?c)}"
                            <#if i == 0>autofocus</#if>
                            <#if maxAttemptsReached?? && maxAttemptsReached>disabled</#if>
                        />
                    </#list>
                </div>

                <input id="emailCode" name="emailCode" type="hidden" value=""
                       aria-invalid="<#if messagesPerField.existsError('emailCode')>true</#if>"
                       <#if maxAttemptsReached?? && maxAttemptsReached>disabled</#if>/>

                <#if messagesPerField.existsError('emailCode')>
                    <span id="input-error-otp-code" class="${properties.kcInputErrorMessageClass!} voco-otp-error"
                          aria-live="polite">
                        ${kcSanitize(messagesPerField.get('emailCode'))?no_esc}
                    </span>
                </#if>
            </div>

            <div class="${properties.kcFormGroupClass!}">
                <div id="kc-form-buttons" class="voco-otp-actions">
                    <#if !(maxAttemptsReached?? && maxAttemptsReached)>
                        <input class="${properties.kcButtonClass!} ${properties.kcButtonPrimaryClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!}"
                               name="login" id="kc-otp-submit" type="submit" value="${msg("doLogIn")}" />
                    </#if>
                    <input class="${properties.kcButtonClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!} voco-btn-secondary"
                           name="resend" type="submit" value="${msg("resendCode")}"/>
                    <input class="${properties.kcButtonClass!} ${properties.kcButtonBlockClass!} ${properties.kcButtonLargeClass!} voco-btn-secondary"
                           name="cancel" type="submit" value="${msg("doCancel")}"/>
                </div>
            </div>
        </form>
    </#if>
</@layout.registrationLayout>
