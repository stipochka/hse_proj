import { useEffect, useState } from 'react'
import { Form, Button, InputNumber, Input, Space, Card, Divider, Spin, Result, Row, Col, Tag } from 'antd'
import { ArrowLeftOutlined, CheckOutlined, CloseOutlined } from '@ant-design/icons'
import { useParams, useNavigate } from 'react-router-dom'
import { activitiesAPI } from '@/shared/api/activities'
import type { Activity, EvaluationRequest } from '@/shared/types'

const ActivityEvaluationPage = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [form] = Form.useForm()
  const [loading, setLoading] = useState(true)
  const [submitting, setSubmitting] = useState(false)
  const [activity, setActivity] = useState<Activity | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [fileUrl, setFileUrl] = useState<string | null>(null)

  useEffect(() => {
    const fetchActivity = async () => {
      if (!id) return
      try {
        const activityId = parseInt(id)
        const [data, fileData] = await Promise.all([
          activitiesAPI.getActivityById(activityId),
          activitiesAPI.getFileUrl(activityId).catch(() => null),
        ])
        setActivity(data)
        if (fileData) {
          setFileUrl(fileData.file_url)
        }
      } catch (err) {
        console.error('Failed to fetch activity:', err)
        setError('Не удалось загрузить активность')
      } finally {
        setLoading(false)
      }
    }

    fetchActivity()
  }, [id])

  const handleEvaluate = async (values: EvaluationRequest) => {
    if (!activity) return

    setSubmitting(true)
    try {
      const evaluation: EvaluationRequest = {
        points: values.points,
        credits: values.credits,
        comment: values.comment,
        reject: false,
      }
      await activitiesAPI.evaluateActivity(activity.id, evaluation)
      navigate('/activities')
    } catch (err) {
      console.error('Failed to evaluate activity:', err)
      setError('Не удалось сохранить оценку')
    } finally {
      setSubmitting(false)
    }
  }

  const handleReject = async () => {
    if (!activity) return

    setSubmitting(true)
    try {
      const comment = form.getFieldValue('comment')
      const evaluation: EvaluationRequest = {
        points: 0,
        comment,
        reject: true,
      }
      await activitiesAPI.evaluateActivity(activity.id, evaluation)
      navigate('/activities')
    } catch (err) {
      console.error('Failed to reject activity:', err)
      setError('Не удалось отклонить активность')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) {
    return (
      <div className="flex justify-center items-center h-96">
        <Spin size="large" />
      </div>
    )
  }

  if (error || !activity) {
    return (
      <Result
        status="error"
        title="Ошибка"
        subTitle={error || 'Активность не найдена'}
        extra={
          <Button type="primary" onClick={() => navigate('/activities')}>
            Вернуться к списку
          </Button>
        }
      />
    )
  }

  return (
    <div>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/activities')}
        className="mb-6"
      >
        Вернуться к списку активностей
      </Button>

      <h1 className="text-3xl font-bold mb-6">Проверка активности</h1>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={12}>
          <Card title="Информация об активности">
            <div className="space-y-4">
              <div>
                <div className="text-gray-600 text-sm">ID Студента</div>
                <div className="text-lg font-semibold">{activity.student_id}</div>
              </div>
              <div>
                <div className="text-gray-600 text-sm">Группа</div>
                <div className="text-lg">{activity.student_group}</div>
              </div>
              <div>
                <div className="text-gray-600 text-sm">Активность</div>
                <div className="text-lg font-semibold">{activity.title}</div>
              </div>
              <div>
                <div className="text-gray-600 text-sm">Описание</div>
                <div className="text-base">{activity.description}</div>
              </div>
              <div>
                <div className="text-gray-600 text-sm">Категория</div>
                <div className="text-lg">{activity.category}</div>
              </div>
              <Divider />
              <div>
                <div className="text-gray-600 text-sm">Статус</div>
                <Tag color={activity.status === 'SUBMITTED' ? 'orange' : 'default'}>
                  {activity.status}
                </Tag>
              </div>
              <div>
                <div className="text-gray-600 text-sm">Создано</div>
                <div className="text-lg">
                  {new Date(activity.created_at).toLocaleDateString('ru-RU', {
                    year: 'numeric',
                    month: 'long',
                    day: 'numeric',
                    hour: '2-digit',
                    minute: '2-digit',
                  })}
                </div>
              </div>

              {activity.evaluation && (
                <>
                  <Divider />
                  <div>
                    <div className="text-gray-600 text-sm font-semibold">Оценка</div>
                    <div>Баллы: {activity.evaluation.points}</div>
                    {activity.evaluation.credits && (
                      <div>З.е.: {activity.evaluation.credits.toFixed(2)}</div>
                    )}
                    {activity.evaluation.comment && (
                      <div>Комментарий: {activity.evaluation.comment}</div>
                    )}
                  </div>
                </>
              )}
            </div>
          </Card>
        </Col>

        <Col xs={24} lg={12}>
          <Card title="Форма оценки">
            <Form
              form={form}
              layout="vertical"
              onFinish={handleEvaluate}
              requiredMark="optional"
            >
              <Form.Item
                label="Баллы"
                name="points"
                rules={[
                  { required: true, message: 'Пожалуйста, введите баллы' },
                  { type: 'number', min: 0, message: 'Баллы не могут быть отрицательными' },
                ]}
              >
                <InputNumber placeholder="Введите баллы" style={{ width: '100%' }} />
              </Form.Item>

              <Form.Item
                label="Зачетные единицы (опционально)"
                name="credits"
              >
                <InputNumber
                  placeholder="Введите кол-во з.е."
                  style={{ width: '100%' }}
                  step={0.1}
                  min={0}
                />
              </Form.Item>

              <Form.Item
                label="Комментарий (опционально)"
                name="comment"
              >
                <Input.TextArea
                  placeholder="Введите комментарий"
                  rows={4}
                />
              </Form.Item>

              <Space style={{ width: '100%' }} direction="vertical">
                <Button
                  type="primary"
                  htmlType="submit"
                  block
                  loading={submitting}
                  icon={<CheckOutlined />}
                >
                  Оценить
                </Button>
                <Button
                  danger
                  block
                  onClick={handleReject}
                  loading={submitting}
                  icon={<CloseOutlined />}
                >
                  Отклонить
                </Button>
              </Space>
            </Form>
          </Card>
        </Col>
      </Row>

      {fileUrl && (
        <div className="mt-6">
          <Card title="PDF документ">
            <iframe
              src={fileUrl}
              title="Activity PDF"
              style={{
                width: '100%',
                height: '600px',
                border: 'none',
                borderRadius: '4px',
              }}
            />
          </Card>
        </div>
      )}
    </div>
  )
}

export default ActivityEvaluationPage

