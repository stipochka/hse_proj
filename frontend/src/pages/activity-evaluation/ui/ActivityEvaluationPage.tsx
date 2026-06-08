import { useEffect, useState } from 'react'
import { Form, Button, InputNumber, Input, Card, Spin, Result, Row, Col, Typography, Divider, Space } from 'antd'
import { ArrowLeftOutlined, CheckOutlined, CloseOutlined } from '@ant-design/icons'
import { useParams, useNavigate } from 'react-router-dom'
import { activitiesAPI } from '@/shared/api/activities'
import { StatusTag } from '@/shared/components/StatusTag'
import { formatDate } from '@/shared/lib/utils'
import type { Activity, EvaluationRequest } from '@/shared/types'

const { Title, Text } = Typography

const ActivityEvaluationPage = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [activity, setActivity] = useState<Activity | null>(null)
  const [fileUrl, setFileUrl] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const load = async () => {
      if (!id) return
      try {
        const actId = parseInt(id)
        const [data, fileData] = await Promise.all([
          activitiesAPI.getActivityById(actId),
          activitiesAPI.getFileUrl(actId).catch(() => null),
        ])
        setActivity(data)
        if (fileData) setFileUrl(fileData.file_url)
      } catch {
        setError('Не удалось загрузить активность')
      } finally {
        setLoading(false)
      }
    }
    load()
  }, [id])

  const submit = async (reject: boolean) => {
    if (!activity) return
    setSubmitting(true)
    try {
      const ev: EvaluationRequest = reject
        ? { points: 0, comment: form.getFieldValue('comment'), reject: true }
        : { ...form.getFieldsValue(), reject: false }
      await activitiesAPI.evaluateActivity(activity.id, ev)
      navigate('/activities')
    } catch {
      setError('Не удалось сохранить оценку')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return (
    <div style={{ display: 'flex', justifyContent: 'center', padding: '80px 0' }}>
      <Spin size="large" />
    </div>
  )

  if (error || !activity) return (
    <Result
      status="error"
      title="Ошибка"
      subTitle={error || 'Активность не найдена'}
      extra={<Button onClick={() => navigate('/activities')}>Вернуться</Button>}
    />
  )

  return (
    <div>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/activities')}
        style={{ marginBottom: 20, paddingLeft: 0 }}
      >
        К списку активностей
      </Button>

      <Title level={3} style={{ marginBottom: 24, fontWeight: 600 }}>Проверка активности</Title>

      <Row gutter={[20, 20]}>
        <Col xs={24} lg={12}>
          <Card
            title="Информация"
            bordered={false}
            style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}
          >
            <Space direction="vertical" size={16} style={{ width: '100%' }}>
              <div>
                <Text type="secondary" style={{ fontSize: 12 }}>Студент</Text>
                <div style={{ fontWeight: 500 }}>{activity.student_name || activity.student_id}</div>
              </div>
              <div>
                <Text type="secondary" style={{ fontSize: 12 }}>Группа</Text>
                <div>{activity.student_group}</div>
              </div>
              <div>
                <Text type="secondary" style={{ fontSize: 12 }}>Активность</Text>
                <div style={{ fontWeight: 500 }}>{activity.title}</div>
              </div>
              {activity.description && (
                <div>
                  <Text type="secondary" style={{ fontSize: 12 }}>Описание</Text>
                  <div>{activity.description}</div>
                </div>
              )}
              <div>
                <Text type="secondary" style={{ fontSize: 12 }}>Категория</Text>
                <div>{activity.category}</div>
              </div>
              <Divider style={{ margin: '4px 0' }} />
              <div style={{ display: 'flex', gap: 24 }}>
                <div>
                  <Text type="secondary" style={{ fontSize: 12 }}>Статус</Text>
                  <div style={{ marginTop: 4 }}><StatusTag status={activity.status} /></div>
                </div>
                <div>
                  <Text type="secondary" style={{ fontSize: 12 }}>Дата подачи</Text>
                  <div>{formatDate(activity.created_at)}</div>
                </div>
              </div>

              {activity.evaluation && (
                <>
                  <Divider style={{ margin: '4px 0' }} />
                  <div>
                    <Text type="secondary" style={{ fontSize: 12 }}>Текущая оценка</Text>
                    <div>
                      Баллы: <strong>{activity.evaluation.points}</strong>
                      {activity.evaluation.credits != null && (
                        <span> · З.е.: <strong>{Number(activity.evaluation.credits).toFixed(2)}</strong></span>
                      )}
                    </div>
                    {activity.evaluation.comment && (
                      <div style={{ color: '#595959', marginTop: 4 }}>{activity.evaluation.comment}</div>
                    )}
                  </div>
                </>
              )}
            </Space>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card
            title="Оценка"
            bordered={false}
            style={{ boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}
          >
            <Form form={form} layout="vertical" requiredMark="optional">
              <Form.Item
                label="Баллы"
                name="points"
                rules={[
                  { required: true, message: 'Введите баллы' },
                  { type: 'number', min: 0, message: 'Не может быть отрицательным' },
                ]}
              >
                <InputNumber style={{ width: '100%' }} placeholder="0" />
              </Form.Item>
              <Form.Item label="Зачётные единицы (необязательно)" name="credits">
                <InputNumber style={{ width: '100%' }} placeholder="0.00" step={0.1} min={0} />
              </Form.Item>
              <Form.Item label="Комментарий (необязательно)" name="comment">
                <Input.TextArea rows={3} placeholder="Комментарий к оценке" style={{ resize: 'none' }} />
              </Form.Item>

              <Space direction="vertical" style={{ width: '100%' }} size={8}>
                <Button
                  type="primary"
                  block
                  icon={<CheckOutlined />}
                  loading={submitting}
                  onClick={() => form.validateFields().then(() => submit(false))}
                >
                  Принять и оценить
                </Button>
                <Button
                  danger
                  block
                  icon={<CloseOutlined />}
                  loading={submitting}
                  onClick={() => submit(true)}
                >
                  Отклонить
                </Button>
              </Space>
            </Form>
          </Card>
        </Col>
      </Row>

      {fileUrl && (
        <Card
          title="Документ"
          bordered={false}
          style={{ marginTop: 20, boxShadow: '0 1px 4px rgba(0,0,0,.08)' }}
        >
          <iframe
            src={fileUrl}
            title="PDF"
            style={{ width: '100%', height: 600, border: 'none', borderRadius: 4 }}
          />
        </Card>
      )}
    </div>
  )
}

export default ActivityEvaluationPage
