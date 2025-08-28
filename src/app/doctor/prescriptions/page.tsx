'use client';

import { useState, useEffect } from 'react';
import { useParams } from 'next/navigation';

interface PrescriptionItem {
  id: number;
  medication_name: string;
  dosage: string;
  frequency: string;
  duration: string;
  instructions: string;
}

interface Prescription {
  id: number;
  appointment_id: number;
  created_by_doctor_id: number;
  prescription_date: string;
  items: PrescriptionItem[];
  notes: string;
  created_at: string;
  updated_at: string;
  appointment: {
    patient: {
      name: string;
      id: number;
    };
    start_time: string;
    end_time: string;
    status: string;
  };
}

interface NewPrescriptionItem {
  medication_name: string;
  dosage: string;
  frequency: string;
  duration: string;
  instructions: string;
}

interface NewPrescription {
  appointment_id: number;
  prescription_date: string;
  items: NewPrescriptionItem[];
  notes: string;
}

export default function DoctorPrescriptionsPage() {
  const params = useParams();
  const appointmentId = params.appointmentId as string;
  
  const [prescriptions, setPrescriptions] = useState<Prescription[]>([]);
  const [selectedPrescription, setSelectedPrescription] = useState<Prescription | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [isCreating, setIsCreating] = useState(false);
  const [showCreateForm, setShowCreateForm] = useState(false);
  
  // 新規作成フォームの状態
  const [newPrescription, setNewPrescription] = useState<NewPrescription>({
    appointment_id: parseInt(appointmentId),
    prescription_date: new Date().toISOString().split('T')[0],
    items: [{ medication_name: '', dosage: '', frequency: '', duration: '', instructions: '' }],
    notes: '',
  });

  useEffect(() => {
    if (appointmentId) {
      loadPrescriptions();
    }
  }, [appointmentId]);

  const loadPrescriptions = async () => {
    try {
      setIsLoading(true);
      const response = await fetch(`/api/v1/appointments/${appointmentId}/prescriptions`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (response.ok) {
        const data = await response.json();
        setPrescriptions(data.prescriptions || []);
      }
    } catch (error) {
      console.error('処方の読み込みに失敗しました:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const createPrescription = async () => {
    try {
      setIsCreating(true);
      const response = await fetch(`/api/v1/appointments/${appointmentId}/prescriptions`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newPrescription),
      });

      if (response.ok) {
        setShowCreateForm(false);
        setNewPrescription({
          appointment_id: parseInt(appointmentId),
          prescription_date: new Date().toISOString().split('T')[0],
          items: [{ medication_name: '', dosage: '', frequency: '', duration: '', instructions: '' }],
          notes: '',
        });
        await loadPrescriptions();
      }
    } catch (error) {
      console.error('処方の作成に失敗しました:', error);
    } finally {
      setIsCreating(false);
    }
  };

  const updatePrescription = async (prescriptionId: number, updatedData: Partial<Prescription>) => {
    try {
      const response = await fetch(`/api/v1/appointments/${appointmentId}/prescriptions/${prescriptionId}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updatedData),
      });

      if (response.ok) {
        await loadPrescriptions();
        setSelectedPrescription(null);
      }
    } catch (error) {
      console.error('処方の更新に失敗しました:', error);
    }
  };

  const deletePrescription = async (prescriptionId: number) => {
    if (!confirm('この処方を削除しますか？')) return;

    try {
      const response = await fetch(`/api/v1/appointments/${appointmentId}/prescriptions/${prescriptionId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      if (response.ok) {
        await loadPrescriptions();
        setSelectedPrescription(null);
      }
    } catch (error) {
      console.error('処方の削除に失敗しました:', error);
    }
  };

  const addPrescriptionItem = () => {
    setNewPrescription(prev => ({
      ...prev,
      items: [...prev.items, { medication_name: '', dosage: '', frequency: '', duration: '', instructions: '' }]
    }));
  };

  const removePrescriptionItem = (index: number) => {
    if (newPrescription.items.length > 1) {
      setNewPrescription(prev => ({
        ...prev,
        items: prev.items.filter((_, i) => i !== index)
      }));
    }
  };

  const updatePrescriptionItem = (index: number, field: keyof PrescriptionItem, value: string) => {
    setNewPrescription(prev => ({
      ...prev,
      items: prev.items.map((item, i) => 
        i === index ? { ...item, [field]: value } : item
      )
    }));
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ja-JP');
  };

  const formatDateTime = (dateString: string) => {
    return new Date(dateString).toLocaleString('ja-JP');
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-6xl mx-auto p-6">
        {/* ヘッダー */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <div className="flex items-center justify-between">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">処方管理</h1>
              <p className="text-gray-600">予約ID: {appointmentId}</p>
            </div>
            <button
              onClick={() => setShowCreateForm(true)}
              className="px-6 py-3 bg-green-600 text-white rounded-lg hover:bg-green-700"
            >
              新規処方作成
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* 処方一覧 */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg shadow-sm">
              <div className="p-6 border-b">
                <h2 className="text-xl font-semibold">処方一覧</h2>
              </div>
              
              {isLoading ? (
                <div className="flex justify-center items-center py-12">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
                </div>
              ) : prescriptions.length === 0 ? (
                <div className="text-center py-12">
                  <div className="text-6xl mb-4">💊</div>
                  <p className="text-gray-500">まだ処方がありません</p>
                </div>
              ) : (
                <div className="divide-y">
                  {prescriptions.map((prescription) => (
                    <div
                      key={prescription.id}
                      className="p-6 hover:bg-gray-50 cursor-pointer"
                      onClick={() => setSelectedPrescription(prescription)}
                    >
                      <div className="flex items-start justify-between">
                        <div className="flex-1">
                          <div className="flex items-center space-x-3 mb-2">
                            <h3 className="text-lg font-medium text-gray-900">
                              処方日: {formatDate(prescription.prescription_date)}
                            </h3>
                            <span className="px-2 py-1 text-xs font-medium rounded-full bg-blue-100 text-blue-800">
                              {prescription.items.length}種類
                            </span>
                          </div>
                          
                          <div className="grid grid-cols-2 gap-4 text-sm text-gray-600">
                            <div>
                              <span className="font-medium">患者:</span> {prescription.appointment.patient.name}
                            </div>
                            <div>
                              <span className="font-medium">予約時間:</span> {formatDateTime(prescription.appointment.start_time)}
                            </div>
                            <div>
                              <span className="font-medium">薬の種類:</span> {prescription.items.length}種類
                            </div>
                            <div>
                              <span className="font-medium">作成日時:</span> {formatDateTime(prescription.created_at)}
                            </div>
                          </div>
                          
                          {prescription.notes && (
                            <div className="mt-3 p-3 bg-blue-50 rounded-lg">
                              <p className="text-sm text-blue-800">
                                <span className="font-medium">注意事項:</span> {prescription.notes}
                              </p>
                            </div>
                          )}
                        </div>
                        
                        <div className="text-right space-y-2">
                          <button className="text-blue-600 hover:text-blue-800 text-sm font-medium block">
                            詳細を見る →
                          </button>
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              deletePrescription(prescription.id);
                            }}
                            className="text-red-600 hover:text-red-800 text-sm font-medium block"
                          >
                            削除
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* 処方詳細・作成フォーム */}
          <div className="lg:col-span-1">
            {showCreateForm ? (
              <div className="bg-white rounded-lg shadow-sm p-6">
                <h3 className="text-lg font-semibold mb-4">新規処方作成</h3>
                
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      処方日
                    </label>
                    <input
                      type="date"
                      value={newPrescription.prescription_date}
                      onChange={(e) => setNewPrescription(prev => ({ ...prev, prescription_date: e.target.value }))}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                  </div>
                  
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      注意事項
                    </label>
                    <textarea
                      value={newPrescription.notes}
                      onChange={(e) => setNewPrescription(prev => ({ ...prev, notes: e.target.value }))}
                      rows={3}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                      placeholder="患者への注意事項を入力してください"
                    />
                  </div>
                  
                  <div>
                    <div className="flex items-center justify-between mb-2">
                      <label className="block text-sm font-medium text-gray-700">
                        処方内容
                      </label>
                      <button
                        onClick={addPrescriptionItem}
                        className="px-3 py-1 bg-blue-600 text-white rounded text-sm hover:bg-blue-700"
                      >
                        + 追加
                      </button>
                    </div>
                    
                    <div className="space-y-3">
                      {newPrescription.items.map((item, index) => (
                        <div key={index} className="p-3 border rounded-lg">
                          <div className="flex items-center justify-between mb-2">
                            <span className="text-sm font-medium">薬 {index + 1}</span>
                            {newPrescription.items.length > 1 && (
                              <button
                                onClick={() => removePrescriptionItem(index)}
                                className="text-red-600 hover:text-red-800 text-sm"
                              >
                                ✕
                              </button>
                            )}
                          </div>
                          
                          <div className="space-y-2">
                            <input
                              type="text"
                              placeholder="薬名"
                              value={item.medication_name}
                              onChange={(e) => updatePrescriptionItem(index, 'medication_name', e.target.value)}
                              className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                            />
                            <input
                              type="text"
                              placeholder="用量"
                              value={item.dosage}
                              onChange={(e) => updatePrescriptionItem(index, 'dosage', e.target.value)}
                              className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                            />
                            <input
                              type="text"
                              placeholder="頻度"
                              value={item.frequency}
                              onChange={(e) => updatePrescriptionItem(index, 'frequency', e.target.value)}
                              className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                            />
                            <input
                              type="text"
                              placeholder="期間"
                              value={item.duration}
                              onChange={(e) => updatePrescriptionItem(index, 'duration', e.target.value)}
                              className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                            />
                            <input
                              type="text"
                              placeholder="指示"
                              value={item.instructions}
                              onChange={(e) => updatePrescriptionItem(index, 'instructions', e.target.value)}
                              className="w-full px-2 py-1 border border-gray-300 rounded text-sm"
                            />
                          </div>
                        </div>
                      ))}
                    </div>
                  </div>
                  
                  <div className="flex space-x-3 pt-4">
                    <button
                      onClick={() => setShowCreateForm(false)}
                      className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-lg hover:bg-gray-50"
                    >
                      キャンセル
                    </button>
                    <button
                      onClick={createPrescription}
                      disabled={isCreating}
                      className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50"
                    >
                      {isCreating ? '作成中...' : '作成'}
                    </button>
                  </div>
                </div>
              </div>
            ) : (
              <div className="bg-white rounded-lg shadow-sm p-6">
                <h3 className="text-lg font-semibold mb-4">処方詳細</h3>
                
                {!selectedPrescription ? (
                  <div className="text-center py-8 text-gray-500">
                    左側の処方を選択してください
                  </div>
                ) : (
                  <div className="space-y-4">
                    <div className="p-4 bg-gray-50 rounded-lg">
                      <h4 className="font-medium mb-2">基本情報</h4>
                      <div className="space-y-2 text-sm">
                        <div>
                          <span className="font-medium">処方日:</span> {formatDate(selectedPrescription.prescription_date)}
                        </div>
                        <div>
                          <span className="font-medium">患者:</span> {selectedPrescription.appointment.patient.name}
                        </div>
                        <div>
                          <span className="font-medium">予約時間:</span> {formatDateTime(selectedPrescription.appointment.start_time)}
                        </div>
                      </div>
                    </div>
                    
                    <div>
                      <h4 className="font-medium mb-2">処方内容</h4>
                      <div className="space-y-3">
                        {selectedPrescription.items.map((item, index) => (
                          <div key={index} className="p-3 border rounded-lg">
                            <div className="font-medium text-gray-900 mb-2">
                              {item.medication_name || `薬 ${index + 1}`}
                            </div>
                            <div className="grid grid-cols-2 gap-2 text-sm text-gray-600">
                              <div>
                                <span className="font-medium">用量:</span> {item.dosage}
                              </div>
                              <div>
                                <span className="font-medium">頻度:</span> {item.frequency}
                              </div>
                              <div>
                                <span className="font-medium">期間:</span> {item.duration}
                              </div>
                            </div>
                            {item.instructions && (
                              <div className="mt-2 text-sm text-gray-600">
                                <span className="font-medium">指示:</span> {item.instructions}
                              </div>
                            )}
                          </div>
                        ))}
                      </div>
                    </div>
                    
                    {selectedPrescription.notes && (
                      <div className="p-4 bg-yellow-50 rounded-lg">
                        <h4 className="font-medium mb-2 text-yellow-800">注意事項</h4>
                        <p className="text-sm text-yellow-700">{selectedPrescription.notes}</p>
                      </div>
                    )}
                    
                    <div className="text-xs text-gray-500 text-center pt-4">
                      最終更新: {formatDateTime(selectedPrescription.updated_at)}
                    </div>
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
