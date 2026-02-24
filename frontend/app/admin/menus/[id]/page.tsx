'use client';

import { useParams } from 'next/navigation';
import MenuEditorPage from '../new/page';

export default function EditMenuPage() {
  const params = useParams();
  const menuId = params.id as string;

  // Reuse the MenuEditorPage for edit mode
  return <MenuEditorPage />;
}
