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

export default function SettingsPaymentGatewayAlipay(props) {
  const { t } = useTranslation();
  const sectionTitle = props.hideSectionTitle
    ? undefined
    : t('支付宝官方支付设置');
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    AlipayAppId: '',
    AlipayPrivateKey: '',
    AlipayPublicKey: '',
    AlipayNotifyUrl: '',
    AlipayReturnUrl: '',
  });
  const formApiRef = useRef(null);

  useEffect(() => {
    if (props.options && formApiRef.current) {
      const currentInputs = {
        AlipayAppId: props.options.AlipayAppId || '',
        AlipayPrivateKey: props.options.AlipayPrivateKey || '',
        AlipayPublicKey: props.options.AlipayPublicKey || '',
        AlipayNotifyUrl: props.options.AlipayNotifyUrl || '',
        AlipayReturnUrl: props.options.AlipayReturnUrl || '',
      };
      setInputs(currentInputs);
      formApiRef.current.setValues(currentInputs);
    }
  }, [props.options]);

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitAlipaySetting = async () => {
    if (props.options.ServerAddress === '') {
      showError(t('请先填写服务器地址'));
      return;
    }

    setLoading(true);
    try {
      // 官方支付允许管理员主动清空字段，因此保存时必须提交空字符串。
      const options = [
        { key: 'AlipayAppId', value: inputs.AlipayAppId || '' },
        { key: 'AlipayPrivateKey', value: inputs.AlipayPrivateKey || '' },
        { key: 'AlipayPublicKey', value: inputs.AlipayPublicKey || '' },
        { key: 'AlipayNotifyUrl', value: inputs.AlipayNotifyUrl || '' },
        { key: 'AlipayReturnUrl', value: inputs.AlipayReturnUrl || '' },
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
                {t('支付宝官方支付配置，支持 PC 网页支付和 H5 支付。')}
                <br />
                {t('回调地址')}：
                {props.options.ServerAddress
                  ? removeTrailingSlash(props.options.ServerAddress)
                  : t('网站地址')}
                /api/alipay/notify
              </>
            }
            style={{ marginBottom: 12 }}
          />
          <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='AlipayAppId'
                label={t('支付宝应用 ID')}
                placeholder={t('例如：2024xxxxxxxxxxxx')}
                extraText={t('支付宝开放平台应用 AppID')}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='AlipayNotifyUrl'
                label={t('异步通知地址')}
                placeholder={t('留空使用默认回调地址')}
                extraText={t('可选，自定义异步通知回调地址')}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.Input
                field='AlipayReturnUrl'
                label={t('支付完成返回地址')}
                placeholder={t('留空使用默认返回地址')}
                extraText={t('可选，用户支付完成后的浏览器跳转地址')}
              />
            </Col>
          </Row>
          <Row
            gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
            style={{ marginTop: 16 }}
          >
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='AlipayPrivateKey'
                label={t('应用私钥')}
                placeholder={t('粘贴支付宝应用私钥，留空将清空当前配置')}
                extraText={t('RSA2 私钥，保存后不会回显')}
                autosize={{ minRows: 3, maxRows: 5 }}
              />
            </Col>
            <Col xs={24} sm={24} md={12} lg={12} xl={12}>
              <Form.TextArea
                field='AlipayPublicKey'
                label={t('支付宝公钥')}
                placeholder={t('粘贴支付宝公钥，留空将清空当前配置')}
                extraText={t('用于验签，从支付宝开放平台获取，保存后不会回显')}
                autosize={{ minRows: 3, maxRows: 5 }}
              />
            </Col>
          </Row>
          <Button onClick={submitAlipaySetting} style={{ marginTop: 16 }}>
            {t('更新支付宝设置')}
          </Button>
        </Form.Section>
      </Form>
    </Spin>
  );
}
