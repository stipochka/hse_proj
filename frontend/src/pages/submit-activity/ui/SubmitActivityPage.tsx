import { useState } from 'react'
import { Form, Input, Button, Upload, Card, Result, Select, Steps, Typography, Alert } from 'antd'
import { UploadOutlined, SendOutlined } from '@ant-design/icons'
import type { UploadFile } from 'antd'
import { activitiesAPI } from '@/shared/api/activities'

const { Title, Text } = Typography

const CATEGORIES = [
  'Научная деятельность',
  'Спорт',
  'Творчество',
  'Волонтёрство',
  'Олимпиады и конкурсы',
  'Прочее',
]

type FormValues = {
  title: string
  category: string
  description: string
}

const SubmitActivityPage = () => {
  const [form] = Form.useForm<FormValues>()
  const [step, setStep] = useState(0)
  const [fileList, setFileList] = useState<UploadFile[]>([])
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (values: FormValues) => {
    if (fileList.length === 0 || !fileList[0].originFileObj) {
      setError('Прикрепите PDF-файл')
      return
    }

    setSubmitting(true)
    setError(null)

    try {
      setStep(1)
      const { activity_id, upload_url } = await activitiesAPI.getUploadUrl({
        title: values.title,
        category: values.category,
        description: values.description,
      })

      setStep(2)
      const file = fileList[0].originFileObj as File
      const resp = await fetch(upload_url, {
        method: 'PUT',
        body: file,
        headers: { 'Content-Type': 'application/pdf' },
      })
      if (!resp.ok) throw new Error(`Ошибка загрузки: ${resp.status}`)

      setStep(3)
      await activitiesAPI.confirmUpload(activity_id)

      setStep(4)
      form.resetFields()
      setFileList([])
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Произошла ошибка при отправке')
      setStep(0)
    } finally {
      setSubmitting(false)
    }
  }

  if (step === 4) {
    return (
      <Result
        status="success"
        title="Активность подана"
        subTitle="Ваша активность отправлена на проверку. Статус можно отследить в разделе «Мои активности»."
        extra={
          <Button type="primary" onClick={() => setStep(0)}>
            Подать ещё одну
          </Button>
        }
      />
    )
  }

  return (
    <div>
      <Title level={3} style={{ marginBottom: 8, fontWeight: 600 }}>Подать активность</Title>
      <Text type="secondary" style={{ display: 'block', marginBottom: 24 }}>
        Загрузите подтверждающий документ в формате PDF
      </Text>

      {submitting && (
        <Card bordered={false} style={{ marginBottom: 24, boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}>
          <Steps
            current={step - 1}
            size="small"
            items={[
              { title: 'Создание' },
              { title: 'Загрузка файла' },
              { title: 'Подтверждение' },
            ]}
          />
        </Card>
      )}

      <Card
        bordered={false}
        style={{ maxWidth: 600, boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          requiredMark="optional"
          style={{ maxWidth: 520 }}
        >
          <Form.Item
            label="Название"
            name="title"
            rules={[{ required: true, message: 'Введите название активности' }]}
          >
            <Input
              size="large"
              placeholder="Например: Участие в хакатоне HSE Cup 2025"
            />
          </Form.Item>

          <Form.Item
            label="Категория"
            name="category"
            rules={[{ required: true, message: 'Выберите категорию' }]}
          >
            <Select
              size="large"
              placeholder="Выберите категорию"
              options={CATEGORIES.map((c) => ({ label: c, value: c }))}
            />
          </Form.Item>

          <Form.Item label="Описание (необязательно)" name="description">
            <Input.TextArea
              rows={3}
              placeholder="Краткое описание достижения"
              style={{ resize: 'none' }}
            />
          </Form.Item>

          <Form.Item
            label="Документ"
            required
            help={
              <Text type="secondary" style={{ fontSize: 12 }}>
                Скан грамоты, сертификата или справки. Только PDF, до 50 МБ.
              </Text>
            }
            style={{ marginBottom: 24 }}
          >
            <Upload
              accept=".pdf"
              maxCount={1}
              fileList={fileList}
              beforeUpload={(file) => {
                if (file.type !== 'application/pdf') {
                  setError('Допускается только PDF')
                  return Upload.LIST_IGNORE
                }
                if (file.size > 50 * 1024 * 1024) {
                  setError('Файл не должен превышать 50 МБ')
                  return Upload.LIST_IGNORE
                }
                setFileList([{ ...file, uid: file.uid, name: file.name, originFileObj: file }])
                setError(null)
                return false
              }}
              onRemove={() => setFileList([])}
            >
              <Button icon={<UploadOutlined />} size="large" style={{ width: '100%' }}>
                Выбрать файл
              </Button>
            </Upload>
          </Form.Item>

          {error && (
            <Alert message={error} type="error" showIcon style={{ marginBottom: 16 }} />
          )}

          <Button
            type="primary"
            htmlType="submit"
            block
            size="large"
            loading={submitting}
            disabled={fileList.length === 0}
            icon={<SendOutlined />}
          >
            Отправить на проверку
          </Button>
        </Form>
      </Card>
    </div>
  )
}

export default SubmitActivityPage
