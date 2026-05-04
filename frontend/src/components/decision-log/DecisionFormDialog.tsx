import { useState, useRef, useEffect, useCallback } from "react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogBody,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Textarea } from "@/components/ui/textarea";
import ConfirmDialog from "@/components/shared/ConfirmDialog";
import { useToast } from "@/components/ui/toast";
import {
  createDecisionLogApi,
  updateDecisionLogApi,
  publishDecisionLogApi,
} from "@/api/decisionLogs";
import { cn } from "@/lib/utils";

const CATEGORY_OPTIONS = [
  { value: "technical", label: "技术" },
  { value: "resource", label: "资源" },
  { value: "requirement", label: "需求" },
  { value: "schedule", label: "进度" },
  { value: "risk", label: "风险" },
  { value: "other", label: "其他" },
] as const;

const MAX_CONTENT_LENGTH = 2000;
const MAX_TAG_LENGTH = 20;

interface DraftData {
  bizKey: string;
  category: string;
  tags: string[];
  content: string;
  logStatus: string;
}

interface DecisionFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  mode: "new" | "edit";
  teamBizKey: string;
  mainBizKey: string;
  draftData?: DraftData;
  loading?: boolean;
  recentTags?: string[];
  onSuccess: () => void;
}

export function DecisionFormDialog({
  open,
  onOpenChange,
  mode,
  teamBizKey,
  mainBizKey,
  draftData,
  loading = false,
  recentTags = [],
  onSuccess,
}: DecisionFormDialogProps) {
  const { addToast } = useToast();

  // Form state
  const [category, setCategory] = useState("");
  const [tags, setTags] = useState<string[]>([]);
  const [content, setContent] = useState("");
  const [tagInput, setTagInput] = useState("");

  // Validation errors
  const [categoryError, setCategoryError] = useState("");
  const [contentError, setContentError] = useState("");
  const [tagError, setTagError] = useState("");

  // Submission state
  const [submitting, setSubmitting] = useState<"draft" | "publish" | null>(
    null,
  );

  // Confirm dialog state
  const [confirmOpen, setConfirmOpen] = useState(false);

  // Recent tags dropdown
  const [showSuggestions, setShowSuggestions] = useState(false);

  const tagInputRef = useRef<HTMLInputElement>(null);
  const selectRef = useRef<HTMLSelectElement>(null);

  // Pre-fill form when draftData changes (edit mode)
  useEffect(() => {
    if (open && mode === "edit" && draftData && !loading) {
      setCategory(draftData.category);
      setTags(draftData.tags);
      setContent(draftData.content);
    } else if (open && mode === "new") {
      setCategory("");
      setTags([]);
      setContent("");
    }
    // Reset errors and submission state on open
    setCategoryError("");
    setContentError("");
    setTagError("");
    setSubmitting(null);
    setTagInput("");
    setShowSuggestions(false);
  }, [open, mode, draftData, loading]);

  // Focus category select when dialog opens
  useEffect(() => {
    if (open && !loading) {
      setTimeout(() => selectRef.current?.focus(), 0);
    }
  }, [open, loading]);

  // Dirty check: compare current state to initial
  const isDirty = useCallback(() => {
    if (mode === "new") {
      return category !== "" || tags.length > 0 || content !== "";
    }
    // Edit mode: compare to draftData
    if (draftData) {
      return (
        category !== draftData.category ||
        JSON.stringify(tags) !== JSON.stringify(draftData.tags) ||
        content !== draftData.content
      );
    }
    return false;
  }, [mode, category, tags, content, draftData]);

  // Close handler with dirty check
  const handleClose = useCallback(() => {
    if (isDirty()) {
      setConfirmOpen(true);
    } else {
      onOpenChange(false);
    }
  }, [isDirty, onOpenChange]);

  // Validate form
  const validate = useCallback((): boolean => {
    let valid = true;
    if (!category) {
      setCategoryError("请选择分类");
      valid = false;
    }
    if (!content.trim()) {
      setContentError("请输入决策内容");
      valid = false;
    }
    return valid;
  }, [category, content]);

  // Add tag helper
  const addTag = useCallback(
    (tag: string) => {
      const trimmed = tag.trim();
      if (!trimmed) return;
      if (trimmed.length > MAX_TAG_LENGTH) {
        setTagError("标签不能超过 20 字符");
        return;
      }
      if (tags.includes(trimmed)) {
        setTagInput("");
        return;
      }
      setTags((prev) => [...prev, trimmed]);
      setTagInput("");
      setTagError("");
      setShowSuggestions(false);
    },
    [tags],
  );

  // Remove tag helper
  const removeTag = useCallback((index: number) => {
    setTags((prev) => prev.filter((_, i) => i !== index));
  }, []);

  // Tag input key handler
  const handleTagKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      if (e.key === "Enter") {
        e.preventDefault();
        addTag(tagInput);
      } else if (e.key === ",") {
        e.preventDefault();
        addTag(tagInput);
      } else if (e.key === "Escape") {
        setShowSuggestions(false);
        e.stopPropagation();
      }
    },
    [addTag, tagInput],
  );

  // Submit: save draft
  const handleSaveDraft = useCallback(async () => {
    if (!validate()) return;
    setSubmitting("draft");
    try {
      if (mode === "edit" && draftData) {
        await updateDecisionLogApi(teamBizKey, mainBizKey, draftData.bizKey, {
          category,
          tags,
          content,
        });
      } else {
        await createDecisionLogApi(teamBizKey, mainBizKey, {
          category,
          tags,
          content,
          logStatus: "draft",
        });
      }
      addToast("草稿已保存", "success");
      onSuccess();
      onOpenChange(false);
    } catch {
      addToast("操作失败，请重试", "error");
    } finally {
      setSubmitting(null);
    }
  }, [
    validate,
    mode,
    draftData,
    teamBizKey,
    mainBizKey,
    category,
    tags,
    content,
    addToast,
    onSuccess,
    onOpenChange,
  ]);

  // Submit: publish
  const handlePublish = useCallback(async () => {
    if (!validate()) return;
    setSubmitting("publish");
    try {
      if (mode === "edit" && draftData) {
        await updateDecisionLogApi(teamBizKey, mainBizKey, draftData.bizKey, {
          category,
          tags,
          content,
        });
        await publishDecisionLogApi(teamBizKey, mainBizKey, draftData.bizKey);
      } else {
        await createDecisionLogApi(teamBizKey, mainBizKey, {
          category,
          tags,
          content,
          logStatus: "published",
        });
      }
      addToast("决策已发布", "success");
      onSuccess();
      onOpenChange(false);
    } catch {
      addToast("操作失败，请重试", "error");
    } finally {
      setSubmitting(null);
    }
  }, [
    validate,
    mode,
    draftData,
    teamBizKey,
    mainBizKey,
    category,
    tags,
    content,
    addToast,
    onSuccess,
    onOpenChange,
  ]);

  // Character counter color
  const counterColor =
    content.length >= MAX_CONTENT_LENGTH
      ? "text-error"
      : content.length > MAX_CONTENT_LENGTH * 0.9
        ? "text-warning"
        : "text-tertiary";

  // Filtered suggestions
  const filteredSuggestions = recentTags
    .filter(
      (t) =>
        !tags.includes(t) && t.toLowerCase().includes(tagInput.toLowerCase()),
    )
    .slice(0, 10);

  const isSubmitting = submitting !== null;

  return (
    <>
      <Dialog open={open} onOpenChange={handleClose}>
        <DialogContent size="sm" aria-describedby={undefined}>
          <DialogHeader>
            <DialogTitle>
              {mode === "new" ? "添加决策" : "编辑决策"}
            </DialogTitle>
          </DialogHeader>
          <DialogBody className="space-y-4">
            {loading ? (
              <div className="flex items-center justify-center py-10 gap-3 text-tertiary text-[13px]">
                <span className="inline-block h-6 w-6 animate-spin rounded-full border-2 border-border border-t-transparent" />
                加载中...
              </div>
            ) : (
              <>
                {/* Category */}
                <div>
                  <label className="block text-[14px] font-medium mb-1.5">
                    分类 <span className="text-error">*</span>
                  </label>
                  <select
                    ref={selectRef}
                    className={cn(
                      "h-10 w-full rounded-md border bg-white px-3 py-2 text-[13px] text-primary shadow-sm transition-all",
                      "focus:outline-none focus:ring-2 focus:ring-primary-200 focus:border-primary-500",
                      categoryError ? "border-error" : "border-border-dark",
                    )}
                    value={category}
                    onChange={(e) => {
                      setCategory(e.target.value);
                      if (e.target.value) setCategoryError("");
                    }}
                  >
                    <option value="">请选择分类</option>
                    {CATEGORY_OPTIONS.map((opt) => (
                      <option key={opt.value} value={opt.value}>
                        {opt.label}
                      </option>
                    ))}
                  </select>
                  {categoryError && (
                    <div className="text-error-text text-xs mt-1" role="alert">
                      {categoryError}
                    </div>
                  )}
                </div>

                {/* Tags */}
                <div>
                  <label className="block text-[14px] font-medium mb-1.5">
                    标签
                  </label>
                  <div
                    className={cn(
                      "flex flex-wrap items-center gap-1 rounded-md border min-h-10 px-2 py-1.5 bg-white shadow-sm transition-all",
                      "focus-within:ring-2 focus-within:ring-primary-200 focus-within:border-primary-500",
                      tagError ? "border-error" : "border-border-dark",
                    )}
                    style={{ position: "relative" }}
                  >
                    {tags.map((tag, idx) => (
                      <span
                        key={tag}
                        className="inline-flex items-center gap-1 rounded-md px-2 py-0.5 bg-bg-alt text-secondary text-xs border border-border"
                      >
                        {tag}
                        <button
                          type="button"
                          onClick={() => removeTag(idx)}
                          className="text-tertiary hover:text-error leading-none"
                          aria-label={`移除标签 ${tag}`}
                        >
                          &times;
                        </button>
                      </span>
                    ))}
                    <input
                      ref={tagInputRef}
                      type="text"
                      className="flex-1 min-w-20 border-none outline-none bg-transparent text-[13px] text-primary py-0.5"
                      placeholder="输入标签，回车添加"
                      value={tagInput}
                      onChange={(e) => {
                        setTagInput(e.target.value);
                        setTagError("");
                        setShowSuggestions(true);
                      }}
                      onFocus={() => setShowSuggestions(true)}
                      onBlur={() =>
                        setTimeout(() => setShowSuggestions(false), 150)
                      }
                      onKeyDown={handleTagKeyDown}
                    />
                    {showSuggestions && filteredSuggestions.length > 0 && (
                      <div className="absolute top-full left-0 right-0 mt-1 bg-white border border-border rounded-md shadow-lg z-10">
                        {filteredSuggestions.map((tag) => (
                          <button
                            key={tag}
                            type="button"
                            className="block w-full text-left px-3 py-2 text-[13px] text-secondary hover:bg-bg-alt"
                            onMouseDown={(e) => e.preventDefault()}
                            onClick={() => addTag(tag)}
                          >
                            {tag}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                  {tagError && (
                    <div className="text-error-text text-xs mt-1" role="alert">
                      {tagError}
                    </div>
                  )}
                </div>

                {/* Content */}
                <div>
                  <label className="block text-[14px] font-medium mb-1.5">
                    决策内容 <span className="text-error">*</span>
                  </label>
                  <Textarea
                    className={cn(contentError && "border-error")}
                    placeholder="请输入决策内容"
                    value={content}
                    onChange={(e) => {
                      setContent(e.target.value);
                      if (e.target.value.trim()) setContentError("");
                    }}
                    aria-label="决策内容"
                  />
                  <div className="flex justify-between items-center mt-1">
                    {contentError ? (
                      <div className="text-error-text text-xs" role="alert">
                        {contentError}
                      </div>
                    ) : (
                      <span />
                    )}
                    <span
                      className={cn("text-xs", counterColor)}
                      aria-live="polite"
                    >
                      {content.length}/{MAX_CONTENT_LENGTH}
                    </span>
                  </div>
                </div>
              </>
            )}
          </DialogBody>
          <DialogFooter>
            <Button
              variant="secondary"
              onClick={handleClose}
              disabled={isSubmitting}
            >
              取消
            </Button>
            <Button
              variant="secondary"
              onClick={handleSaveDraft}
              disabled={isSubmitting || loading}
            >
              {submitting === "draft" && (
                <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-border-dark border-t-transparent mr-1.5" />
              )}
              保存草稿
            </Button>
            <Button
              variant="primary"
              onClick={handlePublish}
              disabled={isSubmitting || loading}
            >
              {submitting === "publish" && (
                <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent mr-1.5" />
              )}
              发布
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <ConfirmDialog
        open={confirmOpen}
        onOpenChange={setConfirmOpen}
        title="确认关闭"
        description="未保存的内容将丢失，确认关闭？"
        onConfirm={() => {
          setConfirmOpen(false);
          onOpenChange(false);
        }}
      />
    </>
  );
}
