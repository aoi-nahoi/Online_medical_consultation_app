'use client';

import { useState, useEffect } from 'react';

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
  doctor: {
    name: string;
    specialty: string;
  };
  appointment: {
    start_time: string;
    end_time: string;
  };
}

export default function PatientPrescriptionsPage() {
  const [prescriptions, setPrescriptions] = useState<Prescription[]>([]);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedPrescription, setSelectedPrescription] = useState<Prescription | null>(null);

  useEffect(() => {
    loadPrescriptions();
  }, []);

  const loadPrescriptions = async () => {
    try {
      setIsLoading(true);
      // 実際の実装では、患者IDに基づいて処方を取得
      const response = await fetch('/api/v1/patient/prescriptions', {
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

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('ja-JP');
  };

  const formatDateTime = (dateString: string) => {
    return new Date(dateString).toLocaleString('ja-JP');
  };

  const getStatusBadge = (prescription: Prescription) => {
    const now = new Date();
    const prescriptionDate = new Date(prescription.prescription_date);
    const daysDiff = Math.floor((now.getTime() - prescriptionDate.getTime()) / (1000 * 60 * 60 * 24));
    
    if (daysDiff <= 7) {
      return <span className="px-2 py-1 text-xs font-medium rounded-full bg-green-100 text-green-800">新しい</span>;
    } else if (daysDiff <= 30) {
      return <span className="px-2 py-1 text-xs font-medium rounded-full bg-yellow-100 text-yellow-800">1ヶ月以内</span>;
    } else {
      return <span className="px-2 py-1 text-xs font-medium rounded-full bg-gray-100 text-gray-800">古い</span>;
    }
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-6xl mx-auto p-6">
        {/* ヘッダー */}
        <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
          <h1 className="text-2xl font-bold text-gray-900">処方履歴</h1>
          <p className="text-gray-600">これまでに処方された薬の履歴を確認できます</p>
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
                            {getStatusBadge(prescription)}
                          </div>
                          
                          <div className="grid grid-cols-2 gap-4 text-sm text-gray-600">
                            <div>
                              <span className="font-medium">医師:</span> {prescription.doctor.name}
                            </div>
                            <div>
                              <span className="font-medium">診療科:</span> {prescription.doctor.specialty}
                            </div>
                            <div>
                              <span className="font-medium">薬の種類:</span> {prescription.items.length}種類
                            </div>
                            <div>
                              <span className="font-medium">処方日時:</span> {formatDateTime(prescription.created_at)}
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
                        
                        <div className="text-right">
                          <button className="text-blue-600 hover:text-blue-800 text-sm font-medium">
                            詳細を見る →
                          </button>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          {/* 処方詳細 */}
          <div className="lg:col-span-1">
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
                        <span className="font-medium">医師:</span> {selectedPrescription.doctor.name}
                      </div>
                      <div>
                        <span className="font-medium">診療科:</span> {selectedPrescription.doctor.specialty}
                      </div>
                    </div>
                  </div>
                  
                  <div>
                    <h4 className="font-medium mb-2">処方内容</h4>
                    <div className="space-y-3">
                      {selectedPrescription.items.map((item) => (
                        <div key={item.id} className="p-3 border rounded-lg">
                          <div className="font-medium text-gray-900 mb-2">
                            {item.medication_name}
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
          </div>
        </div>
      </div>
    </div>
  );
}
