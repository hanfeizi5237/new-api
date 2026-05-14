/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import React, { useEffect, useState, useRef } from 'react';
import { Banner, Button, Form, Row, Col, Spin } from '@douyinfe/semi-ui';
import {
  API,
  removeTrailingSlash,
  showError,
  showSuccess,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';
import { BookOpen } from 'lucide-react';

export default function SettingsPaymentGatewayWxpay(props) {
  const { t } = useTranslation();
  const sectionTitle = props.hideSectionTitle
    ? undefined
    : t('微信官方支付设置');
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    WxpayAppId: '',
    WxpayMchId: '',
    WxpayPrivateKey: '',
    WxpayApiV3Key: '',
    WxpayCertSerial: '',
    WxpayNotifyUrl: '',
    WxpayPublicKey: '',
    WxpayPublicKeyId: '',
  });
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        WxpayAppId: props.options.WxpayAppId || '',
        WxpayMchId: props.options.WxpayMchId || '',
        WxpayPrivateKey: props.options.WxpayPrivateKey || '',
        WxpayApiV3Key: props.options.WxpayApiV3Key || '',
        WxpayCertSerial: props.options.WxpayCertSerial || '',
        WxpayNotifyUrl: props.options.WxpayNotifyUrl || '',
        WxpayPublicKey: props.options.WxpayPublicKey || '',
        WxpayPublicKeyId: props.options.WxpayPublicKeyId || '',
      };
      setInputs(currentInputs);
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitWxpaySetting = async () => {
    if (props.options.ServerAddress === '') {
      showError(t('请先填写服务器地址'));
      return;
    }

    setLoading(true);
    try {
      // 官方支付允许管理员主动清空字段，因此保存时必须提交空字符串。
      const options = [
        { key: 'WxpayAppId', value: inputs.WxpayAppId || '' },
        { key: 'WxpayMchId', value: inputs.WxpayMchId || '' },
        { key: 'WxpayPrivateKey', value: inputs.WxpayPrivateKey || '' },
        { key: 'WxpayApiV3Key', value: inputs.WxpayApiV3Key || '' },
        { key: 'WxpayCertSerial', value: inputs.WxpayCertSerial || '' },
        { key: 'WxpayNotifyUrl', value: inputs.WxpayNotifyUrl || '' },
        { key: 'WxpayPublicKey', value: inputs.WxpayPublicKey || '' },
        { key: 'WxpayPublicKeyId', value: inputs.WxpayPublicKeyId || '' },
      ];

      // 发送请求
      const requestQueue = options.map((opt) =>
        API.put('/api/option/', {
          key: opt.key,
          value: opt.value,
        }),
      );

      const results = await Promise.all(requestQueue);

      // 检查所有请求是否成功
      const errorResults = results.filter((res) => !res.data.success);
      if (errorResults.length > 0) {
        errorResults.forEach((res) => {
          showError(res.data.message);
        });
      } else {
        showSuccess(t('更新成功'));
        props.refresh?.();
      }
    } catch (error) {
      showError(t('更新失败'));
    }
    setLoading(false);
  };

  return (
    <Spin spinning={loading}>
      <Form
        initValues={inputs}
        onValueChange={handleFormChange}
        getFormApi={(api) => (formApiRef.current = api)}
      >
        <Form.Section text={sectionTitle}>
          <Banner
            type='info'
            icon={<BookOpen size={16} />}
            description={
              <>
                {t('微信官方支付配置，支持 Native 扫码支付和 H5 支付。')}
                <br />
                {t('回调地址')}：
                {props.options.ServerAddress
                  ? removeTrailingSlash(props.options.ServerAddress)
                  : t('网站地址')}
                /api/wxpay/notify
              </>
            }
            style={{ marginBottom: 12 }}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WxpayAppId'
                label={t('微信应用 ID')}
                placeholder={t('例如：wx1xxxxxxxxxxxxxxx')}
                extraText={t('微信公众号或移动应用 AppID')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WxpayMchId'
                label={t('微信商户号')}
                placeholder={t('例如：1234567890')}
                extraText={t('微信支付商户号')}
              />
            </Col>
            <Col xs={24} sm={24} md={8} lg={8} xl={8}>
              <Form.Input
                field='WxpayCertSerial'
                label={t('证书序列号')}
                placeholder={t('例如：343459ACC5A6C...')}
                extraText={t('商户 API 证书序列号')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='WxpayPrivateKey'
                label={t('商户私钥')}
                placeholder={t('粘贴微信支付商户私钥，留空将清空当前配置')}
                extraText={t('APIv3 私钥，保存后不会回显')}
                autosize={{ minRows: 3, maxRows: 5 }}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='WxpayApiV3Key'
                label={t('APIv3 密钥')}
                placeholder={t('粘贴微信支付 APIv3 密钥，留空将清空当前配置')}
                extraText={t('用于回调解密和验签，保存后不会回显')}
                autosize={{ minRows: 3, maxRows: 5 }}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='WxpayNotifyUrl'
                label={t('异步通知地址')}
                placeholder={t('留空使用默认回调地址')}
                extraText={t('可选，自定义异步通知回调地址')}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='WxpayPublicKeyId'
                label={t('微信支付公钥 ID')}
                placeholder={t('例如：PUB_KEY_ID_011423...')}
                extraText={t('微信支付公钥模式的公钥 ID')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={24} lg={24} xl={24}>
              <Form.TextArea
                field='WxpayPublicKey'
                label={t('微信支付公钥')}
                placeholder={t('粘贴微信支付公钥，留空将清空当前配置')}
                extraText={t('用于微信支付公钥模式验签，保存后不会回显')}
                autosize={{ minRows: 3, maxRows: 5 }}
              />
            </Col>
          </Row>
          <Button onClick={submitWxpaySetting} style={{ marginTop: 16 }}>
            {t('更新微信支付设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
