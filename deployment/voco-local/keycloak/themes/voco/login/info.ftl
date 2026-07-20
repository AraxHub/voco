<#import "template.ftl" as layout>
<@layout.registrationLayout displayMessage=false; section>
    <#if section = "header">
        <#if messageHeader??>
            ${kcSanitize(msg("${messageHeader}"))?no_esc}
        <#else>
            ${message.summary}
        </#if>
    <#elseif section = "form">
        <div id="kc-info-message" class="voco-info-panel">
            <p class="instruction voco-info-text">
                ${message.summary}
                <#if requiredActions??>
                    <#list requiredActions>
                        : <b><#items as reqActionItem>${kcSanitize(msg("requiredAction.${reqActionItem}"))?no_esc}<#sep>, </#items></b>
                    </#list>
                </#if>
            </p>

            <#if skipLink??>
            <#else>
                <#if pageRedirectUri?has_content>
                    <a class="pf-v5-c-button pf-m-primary pf-m-block voco-cta" href="${pageRedirectUri}">${msg("backToApplication")}</a>
                <#elseif actionUri?has_content>
                    <a class="pf-v5-c-button pf-m-primary pf-m-block voco-cta" href="${actionUri}">${msg("proceedWithAction")}</a>
                <#elseif (client.baseUrl)?has_content>
                    <a class="pf-v5-c-button pf-m-primary pf-m-block voco-cta" href="${client.baseUrl}">${msg("backToApplication")}</a>
                <#else>
                    <a class="pf-v5-c-button pf-m-primary pf-m-block voco-cta" href="${url.loginUrl}">${msg("backToLogin")}</a>
                </#if>
            </#if>
        </div>
    </#if>
</@layout.registrationLayout>
