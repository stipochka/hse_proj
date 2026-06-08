import { useState } from 'react'
import { Form, Input, Button, Upload, Card, Result, Select, Steps } from 'antd'
import { UploadOutlined, CheckCircleOutlined } from '@ant-design/icons'
import type { UploadFile } from 'antd'
import { activitiesAPI } from '@/shared/api/activities'

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
      // 1. Получаем presigned URL и создаём активность
      setStep(1)
      const { activity_id, upload_url } = await activitiesAPI.getUploadUrl({
        title: values.title,
        category: values.category,
        description: values.description,
      })

      // 2. Загружаем PDF напрямую в S3
      setStep(2)
      const file = fileList[0].originFileObj as File
      const uploadResp = await fetch(upload_url, {
        method: 'PUT',
        body: file,
        headers: { 'Content-Type': 'application/pdf' },
      })
      if (!uploadResp.ok) {
        throw new Error(`Ошибка загрузки файла: ${uploadResp.status}`)
      }

      // 3. Подтверждаем загрузку
      setStep(3)
      await activitiesAPI.confirmUpload(activity_id)

      setStep(4)
      form.resetFields()
      setFileList([])
    } catch (err: any) {
      setError(err?.message || 'Произошла ошибка при отправке')
      setStep(0)
    } finally {
      setSubmitting(false)
    }
  }

  if (step === 4) {
    return (
      <Result
        icon={<CheckCircleOutlined style={{ color: '#52c41a' }} />}
        status="success"
        title="Активность подана!"
        subTitle="Ваша активность отправлена на проверку. Вы можете отследить статус в разделе «Мои активности»."
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
      <h1 className="text-3xl font-bold mb-6">Подать активность</h1>

      {submitting && (
        <Steps
          current={step - 1}
          className="mb-6"
          items={[
            { title: 'Создание' },
            { title: 'Загрузка файла' },
            { title: 'Подтверждение' },
          ]}
        />
      )}

      <Card style={{ maxWidth: 640 }}>
        <Form form={form} layout="vertical" onFinish={handleSubmit} requiredMark="optional">
          <Form.Item
            label="Название"
            name="title"
            rules={[{ required: true, message: 'Введите название активности' }]}
          >
            <Input placeholder="Например: Участие в хакатоне HSE Cup 2025" />
          </Form.Item>

          <Form.Item
            label="Категория"
            name="category"
            rules={[{ required: true, message: 'Выберите категорию' }]}
          >
            <Select
              placeholder="Выберите категорию"
              options={CATEGORIES.map((c) => ({ label: c, value: c }))}
            />
          </Form.Item>

          <Form.Item label="Описание" name="description">
            <Input.TextArea
              placeholder="Краткое описание достижения"
              rows={3}
            />
          </Form.Item>

          <Form.Item
            label="PDF-документ (подтверждение)"
            required
            help="Загрузите скан грамоты, сертификата или справки"
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
                return false // не отправляем через antd, делаем сами
              }}
              onRemove={() => setFileList([])}
            >
              <Button icon={<UploadOutlined />}>Выбрать файл</Button>
            </Upload>
          </Form.Item>

          {error && (
            <div className="mb-4 text-red-500 text-sm">{error}</div>
          )}

          <Button
            type="primary"
            htmlType="submit"
            block
            loading={submitting}
            disabled={fileList.length === 0}
          >
            Отправить на проверку
          </Button>
        </Form>
      </Card>
    </div>
  )
}

export default SubmitActivityPage
